package internal

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/googleapis/gax-go/v2"
	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/flow"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"os"
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

type PipelineJobClient interface {
	GetPipelineJob(
		ctx context.Context,
		req *aiplatformpb.GetPipelineJobRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.PipelineJob, error)
}

type VaiEventSource struct {
	RunsSubscription *pubsub.Subscription
	ProviderConfig   VAIProviderConfig
	Logger           logr.Logger
	out              chan any
}

type VaiResource struct {
	Labels map[string]string `json:"labels"`
}

type VaiLogEntry struct {
	Labels   map[string]string `json:"labels"`
	Resource VaiResource       `json:"resource"`
}

func NewVaiEventSource(
	ctx context.Context,
	config VAIProviderConfig,
	pipelineJobClient PipelineJobClient,
) (*VaiEventSource, error) {
	logger := common.LoggerFromContext(ctx)

	pubSubClient, err := pubsub.NewClient(ctx, config.Parameters.VaiProject)
	if err != nil {
		logger.Error(err, "failed to create pubsub client", "project", config.Parameters.VaiProject)
		return nil, err
	}
	runsSubscription := pubSubClient.Subscription(config.Parameters.EventsourcePipelineEventsSubscription)

	vaiEventDataSource := &VaiEventSource{
		ProviderConfig:   config,
		RunsSubscription: runsSubscription,
		Logger:           logger,
		out:              make(chan any),
	}

	go func() {
		if err := vaiEventDataSource.subscribe(ctx, pipelineJobClient); err != nil {
			logger.Error(err, "Failed to subscribe", "subscription", config.Parameters.EventsourcePipelineEventsSubscription)
			os.Exit(1)
		}
	}()

	return vaiEventDataSource, nil
}

func (es *VaiEventSource) Via(operator streams.Flow) streams.Flow {
	flow.DoStream(es, operator)
	return operator
}

func (es *VaiEventSource) Out() <-chan any {
	return es.out
}

func (es *VaiEventSource) subscribe(ctx context.Context, pipelineJobClient PipelineJobClient) error {
	es.Logger.Info("subscribing to pubsub...")

	err := es.RunsSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		es.Logger.Info(fmt.Sprintf("message received from Pub/Sub with ID: %s", m.ID))
		logEntry := VaiLogEntry{}
		err := json.Unmarshal(m.Data, &logEntry)
		if err != nil {
			es.Logger.Error(err, "failed to unmarshal Pub/Sub message")
			m.Nack()
			return
		}
		es.Logger.Info(fmt.Sprintf("%+v", logEntry))

		pipelineJobId, ok := logEntry.Resource.Labels["pipeline_job_id"]
		if !ok {
			es.Logger.Error(err, fmt.Sprintf("logEntry did not contain pipeline_job_id %+v", logEntry))
			m.Nack()
			return
		}

		event := es.runCompletionEventDataForRun(ctx, pipelineJobId, pipelineJobClient)
		if event == nil {
			es.Logger.Error(err, fmt.Sprintf("failed to convert to run completion event data %s", pipelineJobId))
			m.Nack()
			return
		}

		select {
		case es.out <- StreamMessage{
			RunCompletionEventData: *event,
			OnCompleteHandlers: OnCompleteHandlers{
				OnSuccessHandler: func() { m.Ack() },
				OnFailureHandler: func() { m.Nack() },
			},
		}:
		case <-ctx.Done():
			es.Logger.Info("stopped reading from pubsub")
			return
		}
	})

	if err != nil {
		es.Logger.Error(err, "failed to read from pubsub")
		return err
	}

	return nil
}

func (es *VaiEventSource) runCompletionEventDataForRun(
	ctx context.Context,
	runId string,
	pipelineJobClient PipelineJobClient,
) *common.RunCompletionEventData {
	job, err := pipelineJobClient.GetPipelineJob(ctx, &aiplatformpb.GetPipelineJobRequest{
		Name: es.ProviderConfig.pipelineJobName(runId),
	})
	if err != nil {
		es.Logger.Error(err, "could not fetch pipeline job")
		return nil
	}
	if job == nil {
		es.Logger.Error(nil, "expected pipeline job not found", "run-id", runId)
		return nil
	}

	return es.toRunCompletionEventData(job, runId)
}

func (es *VaiEventSource) toRunCompletionEventData(job *aiplatformpb.PipelineJob, runId string) *common.RunCompletionEventData {
	runCompletionStatus, completed := runCompletionStatus(job)

	if !completed {
		es.Logger.Error(nil, "expected pipeline job to have finished", "run-id", runId)
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
		Provider:              es.ProviderConfig.Name,
	}
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
