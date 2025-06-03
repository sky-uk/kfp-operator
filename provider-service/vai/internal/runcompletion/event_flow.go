package runcompletion

import (
	"context"
	"errors"
	"time"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	PushedModelArtifactType        = "tfx.PushedModel"
	ModelPushedMetadataProperty    = "pushed"
	ModelPushedMetadataValue       = 1
	ModelPushedDestinationProperty = "pushed_destination"
	PipelineJobNotFinishedErr      = "expected pipeline job to have finished"
)

type EventFlow struct {
	ProviderConfig    config.VAIProviderConfig
	PipelineJobClient client.PipelineJobClient
	in                chan StreamMessage[string]
	out               chan StreamMessage[*common.RunCompletionEventData]
	errorOut          chan error
}

func (vef *EventFlow) In() chan<- StreamMessage[string] {
	return vef.in
}

func (vef *EventFlow) Out() <-chan StreamMessage[*common.RunCompletionEventData] {
	return vef.out
}

func (vef *EventFlow) ErrOut() <-chan error {
	return vef.errorOut
}

func (vef *EventFlow) From(outlet streams.Outlet[StreamMessage[string]]) streams.Flow[StreamMessage[string], StreamMessage[*common.RunCompletionEventData], error] {
	go func() {
		for message := range outlet.Out() {
			vef.In() <- message
		}
	}()
	return vef
}

func (vef *EventFlow) To(inlet streams.Inlet[StreamMessage[*common.RunCompletionEventData]]) {
	go func() {
		for message := range vef.out {
			inlet.In() <- message
		}
	}()
}

func (vef *EventFlow) Error(inlet streams.Inlet[error]) {
	for errorMessage := range vef.errorOut {
		inlet.In() <- errorMessage
	}
}

func NewEventFlow(config *config.VAIProviderConfig, pipelineJobClient *aiplatform.PipelineClient) *EventFlow {
	vaiEventFlow := EventFlow{
		ProviderConfig:    *config,
		PipelineJobClient: pipelineJobClient,
		in:                make(chan StreamMessage[string]),
		out:               make(chan StreamMessage[*common.RunCompletionEventData]),
		errorOut:          make(chan error),
	}

	return &vaiEventFlow
}

func (vef *EventFlow) Start(ctx context.Context) {
	go func() {
		logger := common.LoggerFromContext(ctx)
		for msg := range vef.in {
			logger.Info("in VAI flow - received message", "message", msg.Message)
			runCompletionEventData, err := vef.runCompletionEventDataForRun(ctx, msg.Message)
			if err != nil {
				if status.Code(err) == codes.NotFound {
					logger.Info("pipeline job not found", "run-id", msg.Message)
					msg.OnUnrecoverableFailureHandler()
					vef.errorOut <- err
				} else {
					logger.Info("error retrieving job", "run-id", msg.Message)
					msg.OnRecoverableFailureHandler()
					vef.errorOut <- err
				}
			} else {
				vef.out <- StreamMessage[*common.RunCompletionEventData]{
					Message:            runCompletionEventData,
					OnCompleteHandlers: msg.OnCompleteHandlers,
				}
			}
		}
	}()
}

func (vef *EventFlow) runCompletionEventDataForRun(ctx context.Context, runId string) (*common.RunCompletionEventData, error) {
	job, err := vef.PipelineJobClient.GetPipelineJob(ctx, &aiplatformpb.GetPipelineJobRequest{
		Name: vef.ProviderConfig.PipelineJobName(runId),
	})
	if err != nil {
		common.LoggerFromContext(ctx).Error(err, "failed to fetch pipeline job")
		return nil, err
	}
	return vef.toRunCompletionEventData(ctx, job, runId)
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

func (vef *EventFlow) toRunCompletionEventData(ctx context.Context, job *aiplatformpb.PipelineJob, runId string) (*common.RunCompletionEventData, error) {
	runCompletionStatus, completed := runCompletionStatus(job)

	if !completed {
		err := errors.New(PipelineJobNotFinishedErr)
		common.LoggerFromContext(ctx).Error(err, "run-id", runId)
		return nil, err
	}

	var pipelineName common.NamespacedName

	pipelineName.Name = job.Labels[label.PipelineName]
	if pipelineNamespace, ok := job.Labels[label.PipelineNamespace]; ok {
		pipelineName.Namespace = pipelineNamespace
	}

	runName := common.NamespacedName{
		Name:      job.Labels[label.RunName],
		Namespace: job.Labels[label.RunNamespace],
	}

	runConfigurationName := common.NamespacedName{
		Name:      job.Labels[label.RunConfigurationName],
		Namespace: job.Labels[label.RunConfigurationNamespace],
	}
	
	var runStartTime *time.Time
	if job.StartTime != nil && !job.StartTime.AsTime().IsZero() {
		startTime := job.StartTime.AsTime().UTC()
		runStartTime = &startTime
	}

	var runEndTime *time.Time
	if job.EndTime != nil && !job.EndTime.AsTime().IsZero() {
		endTime := job.EndTime.AsTime().UTC()
		runEndTime = &endTime
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
		RunStartTime:          runStartTime,
		RunEndTime:            runEndTime,
	}, nil
}
