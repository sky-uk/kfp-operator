package internal

import (
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"context"
	"github.com/go-logr/logr"
	"github.com/googleapis/gax-go/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams"
)

const (
	PushedModelArtifactType        = "tfx.PushedModel"
	ModelPushedMetadataProperty    = "pushed"
	ModelPushedMetadataValue       = 1
	ModelPushedDestinationProperty = "pushed_destination"
)

var labels = struct {
	PipelineName              string
	PipelineNamespace         string
	PipelineVersion           string
	RunConfigurationName      string
	RunConfigurationNamespace string
	RunName                   string
	RunNamespace              string
}{
	PipelineName:              "pipeline-name",
	PipelineNamespace:         "pipeline-namespace",
	PipelineVersion:           "pipeline-version",
	RunConfigurationName:      "runconfiguration-name",
	RunConfigurationNamespace: "runconfiguration-namespace",
	RunName:                   "run-name",
	RunNamespace:              "run-namespace",
}

type VaiEventFlow struct {
	ProviderConfig    VAIProviderConfig
	PipelineJobClient PipelineJobClient
	Logger            logr.Logger
	context           context.Context
	in                chan StreamMessage[string]
	out               chan StreamMessage[*common.RunCompletionEventData]
	errorOut          chan error
}

func (vef *VaiEventFlow) In() chan<- StreamMessage[string] {
	return vef.in
}

func (vef *VaiEventFlow) Out() <-chan StreamMessage[*common.RunCompletionEventData] {
	return vef.out
}

func (vef *VaiEventFlow) ErrOut() <-chan error {
	return vef.errorOut
}

func (vef *VaiEventFlow) From(outlet streams.Outlet[StreamMessage[string]]) streams.Flow[StreamMessage[string], StreamMessage[*common.RunCompletionEventData], error] {
	go func() {
		for message := range outlet.Out() {
			vef.In() <- message
		}
	}()
	return vef
}

func (vef *VaiEventFlow) To(inlet streams.Inlet[StreamMessage[*common.RunCompletionEventData]]) {
	go func() {
		for message := range vef.out {
			inlet.In() <- message
		}
	}()
}

func (vef *VaiEventFlow) Error(inlet streams.Inlet[error]) {
	for errorMessage := range vef.errorOut {
		inlet.In() <- errorMessage
	}
}

func NewVaiEventFlow(ctx context.Context, config *VAIProviderConfig, pipelineJobClient *aiplatform.PipelineClient) streams.Flow[StreamMessage[string], StreamMessage[*common.RunCompletionEventData], error] {
	logger := common.LoggerFromContext(ctx)

	vaiEventFlow := VaiEventFlow{
		ProviderConfig:    *config,
		PipelineJobClient: pipelineJobClient,
		Logger:            logger,
		context:           ctx,
	}

	go func() {
		for msg := range vaiEventFlow.in {
			runCompletionEvent := vaiEventFlow.runCompletionEventDataForRun(msg.Message)
			vaiEventFlow.out <- StreamMessage[*common.RunCompletionEventData]{
				Message:            runCompletionEvent,
				OnCompleteHandlers: msg.OnCompleteHandlers,
			}
		}
	}()

	return &vaiEventFlow
}

type PipelineJobClient interface {
	GetPipelineJob(
		ctx context.Context,
		req *aiplatformpb.GetPipelineJobRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.PipelineJob, error)
}

func (vef *VaiEventFlow) runCompletionEventDataForRun(runId string) *common.RunCompletionEventData {
	job, err := vef.PipelineJobClient.GetPipelineJob(vef.context, &aiplatformpb.GetPipelineJobRequest{
		Name: vef.ProviderConfig.pipelineJobName(runId),
	})
	if err != nil {
		vef.Logger.Error(err, "could not fetch pipeline job")
		return nil
	}
	if job == nil {
		vef.Logger.Error(nil, "expected pipeline job not found", "run-id", runId)
		return nil
	}

	return vef.toRunCompletionEventData(job, runId)
}

func runCompletionStatus(job *aiplatformpb.PipelineJob) (common.RunCompletionStatus, bool) {
	switch job.State {
	case aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED:
		return common.RunCompletionStatuses.Succeeded, true
	case aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED:
		return common.RunCompletionStatuses.Failed, true
	default:
		return "", false
	}
}

func artifactsFilterData(job *aiplatformpb.PipelineJob) []common.PipelineComponent {
	componentCompletions := make([]common.PipelineComponent, 0, len(job.GetJobDetail().GetTaskDetails()))

	for _, task := range job.GetJobDetail().GetTaskDetails() {
		componentArtifactDetails := make([]common.ComponentArtifact, 0, len(task.GetOutputs()))

		for outputName, output := range task.GetOutputs() {
			artifacts := make([]common.ComponentArtifactInstance, 0)
			for _, artifact := range output.GetArtifacts() {
				metadata := artifact.Metadata.AsMap()
				artifacts = append(artifacts, common.ComponentArtifactInstance{
					Uri:      artifact.Uri,
					Metadata: metadata,
				})
			}
			componentArtifactDetails = append(componentArtifactDetails, common.ComponentArtifact{Name: outputName, Artifacts: artifacts})
		}
		componentCompletions = append(componentCompletions, common.PipelineComponent{Name: task.TaskName, ComponentArtifacts: componentArtifactDetails})
	}

	return componentCompletions
}

func modelServingArtifactsForJob(job *aiplatformpb.PipelineJob) []common.Artifact {
	servingModelArtifacts := []common.Artifact{}
	for _, task := range job.GetJobDetail().GetTaskDetails() {
		for name, output := range task.GetOutputs() {
			for _, artifact := range output.GetArtifacts() {
				if artifact.SchemaTitle != PushedModelArtifactType {
					continue
				}

				properties := artifact.Metadata.AsMap()

				pushedProperty, hasPushed := properties[ModelPushedMetadataProperty]
				if !hasPushed {
					continue
				}

				pushed, isFloat := pushedProperty.(float64)
				if !isFloat || pushed != ModelPushedMetadataValue {
					continue
				}

				pushedDestinationProperty, hasPushedDestination := properties[ModelPushedDestinationProperty]
				if !hasPushedDestination {
					continue
				}

				pushedDestination, isString := pushedDestinationProperty.(string)
				if !isString {
					continue
				}

				servingModelArtifacts = append(servingModelArtifacts, common.Artifact{Name: name, Location: pushedDestination})
			}
		}
	}

	return servingModelArtifacts
}

func (vef *VaiEventFlow) toRunCompletionEventData(job *aiplatformpb.PipelineJob, runId string) *common.RunCompletionEventData {
	runCompletionStatus, completed := runCompletionStatus(job)

	if !completed {
		vef.Logger.Error(nil, "expected pipeline job to have finished", "run-id", runId)
		return nil
	}

	var pipelineName common.NamespacedName

	pipelineName.Name = job.Labels[labels.PipelineName]
	if pipelineNamespace, ok := job.Labels[labels.PipelineNamespace]; ok {
		pipelineName.Namespace = pipelineNamespace
	}

	runName := common.NamespacedName{
		Name:      job.Labels[labels.RunName],
		Namespace: job.Labels[labels.RunNamespace],
	}

	runConfigurationName := common.NamespacedName{
		Name:      job.Labels[labels.RunConfigurationName],
		Namespace: job.Labels[labels.RunConfigurationNamespace],
	}

	return &common.RunCompletionEventData{
		Status:                runCompletionStatus,
		PipelineName:          pipelineName,
		RunConfigurationName:  runConfigurationName.NonEmptyPtr(),
		RunName:               runName.NonEmptyPtr(),
		RunId:                 runId,
		ServingModelArtifacts: modelServingArtifactsForJob(job),
		PipelineComponents:    artifactsFilterData(job),
		Provider:              vef.ProviderConfig.Name,
	}
}

func (vef *VaiEventFlow) ToRunCompletionEventData(message StreamMessage[string]) StreamMessage[*common.RunCompletionEventData] {
	event := vef.runCompletionEventDataForRun(message.Message)
	if event == nil {
		vef.Logger.Info("failed to convert to run completion event data", "jobId", message.Message)
		message.OnFailureHandler()
		return StreamMessage[*common.RunCompletionEventData]{}
	}

	return StreamMessage[*common.RunCompletionEventData]{
		Message:            event,
		OnCompleteHandlers: message.OnCompleteHandlers,
	}
}
