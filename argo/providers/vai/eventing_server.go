package vai

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-logr/logr"
	"github.com/googleapis/gax-go/v2"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/base/generic"
	aiplatformpb "google.golang.org/genproto/googleapis/cloud/aiplatform/v1"
	"gopkg.in/yaml.v2"
)

const (
	PushedModelArtifactType     = "tfx.PushedModel"
	ModelPushedMetadataProperty = "pushed"
	ModelPushedMetadataValue    = 1
)

type PipelineJobClient interface {
	GetPipelineJob(ctx context.Context, req *aiplatformpb.GetPipelineJobRequest, opts ...gax.CallOption) (*aiplatformpb.PipelineJob, error)
}

type VaiEventingServer struct {
	generic.UnimplementedEventingServer
	ProviderConfig    VAIProviderConfig
	Logger            logr.Logger
	RunsSubscription  *pubsub.Subscription
	PipelineJobClient PipelineJobClient
}

type VaiEventSourceConfig struct {
}

func (es *VaiEventingServer) StartEventSource(source *generic.EventSource, stream generic.Eventing_StartEventSourceServer) error {
	eventsourceConfig := VaiEventSourceConfig{}

	if err := yaml.Unmarshal(source.Config, &eventsourceConfig); err != nil {
		es.Logger.Error(err, "failed to parse event source configuration")
		return err
	}

	es.Logger.Info("starting stream", "eventsource", eventsourceConfig)

	err := es.RunsSubscription.Receive(stream.Context(), func(ctx context.Context, m *pubsub.Message) {
		run := VAIRun{}
		err := json.Unmarshal(m.Data, &run)
		if err != nil {
			es.Logger.Error(err, "failed to unmarshal Pub/Sub message")
			m.Nack()
			return
		}

		event, err := es.runCompletionEventForRun(stream.Context(), run.RunId)
		if err != nil || event == nil {
			es.Logger.Error(err, "failed to fetch pipeline job")
			m.Nack()
			return
		}

		jsonPayload, err := json.Marshal(event)
		if err != nil {
			es.Logger.Error(err, "failed to marshal event")
			m.Nack()
			return
		}

		es.Logger.V(1).Info("sending run completion event", "event", event)
		if err = stream.Send(&generic.Event{
			Name:    RunCompletionEventName,
			Payload: jsonPayload,
		}); err != nil {
			es.Logger.Error(err, "failed to send event")
			m.Nack()
			return
		}

		m.Ack()
	})

	if err != nil {
		return err
	}

	return nil
}

func (es *VaiEventingServer) runCompletionEventForRun(ctx context.Context, runId string) (*RunCompletionEvent, error) {
	job, err := es.PipelineJobClient.GetPipelineJob(ctx, &aiplatformpb.GetPipelineJobRequest{
		Name: es.ProviderConfig.pipelineJobName(runId),
	})
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("pipeline job not found")
	}

	return toRunCompletionEvent(job), nil
}

func modelServingArtifactsForJob(job *aiplatformpb.PipelineJob) []ServingModelArtifact {
	var servingModelArtifacts []ServingModelArtifact
	for _, task := range job.GetJobDetail().GetTaskDetails() {
		for name, output := range task.GetOutputs() {
			for _, artifact := range output.GetArtifacts() {
				pushed := artifact.Metadata.AsMap()[ModelPushedMetadataProperty].(float64) == ModelPushedMetadataValue
				if artifact.SchemaTitle == PushedModelArtifactType && pushed {
					servingModelArtifacts = append(servingModelArtifacts, ServingModelArtifact{Name: name, Location: artifact.GetUri()})
				}
			}
		}
	}

	return servingModelArtifacts
}

func toRunCompletionEvent(job *aiplatformpb.PipelineJob) *RunCompletionEvent {
	runCompletionStatus, completed := runCompletionStatus(job)

	if !completed {
		return nil
	}

	return &RunCompletionEvent{
		Status:                runCompletionStatus,
		PipelineName:          job.Labels[labels.PipelineName],
		RunConfigurationName:  job.Labels[labels.RunConfiguration],
		ServingModelArtifacts: modelServingArtifactsForJob(job),
	}
}

func runCompletionStatus(job *aiplatformpb.PipelineJob) (RunCompletionStatus, bool) {
	switch job.State {
	case aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED:
		return Succeeded, true
	case aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED:
		return Failed, true
	default:
		return "", false
	}
}
