package vai

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-events/eventsources/sources/generic"
	"github.com/go-logr/logr"
	"github.com/googleapis/gax-go/v2"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"gopkg.in/yaml.v2"
)

const (
	PushedModelArtifactType        = "tfx.PushedModel"
	ModelPushedMetadataProperty    = "pushed"
	ModelPushedMetadataValue       = 1
	ModelPushedDestinationProperty = "pushed_destination"
)

type PipelineJobClient interface {
	GetPipelineJob(ctx context.Context, req *aiplatformpb.GetPipelineJobRequest, opts ...gax.CallOption) (*aiplatformpb.PipelineJob, error)
}

type VaiEventingServer struct {
	generic.UnimplementedEventingServer
	base.K8sApi
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
		logEntry := VAILogEntry{}
		err := json.Unmarshal(m.Data, &logEntry)
		if err != nil {
			es.Logger.Error(err, "failed to unmarshal Pub/Sub message")
			m.Nack()
			return
		}

		pipelineJobId, ok := logEntry.Resource.Labels["pipeline_job_id"]
		if !ok {
			es.Logger.Error(err, fmt.Sprintf("logEntry did not contain pipeline_job_id %+v", logEntry))
			m.Nack()
			return
		}

		event := es.runCompletionEventDataForRun(stream.Context(), pipelineJobId)
		if event == nil {
			es.Logger.Error(err, fmt.Sprintf("failed to convert to run completion event data %s", pipelineJobId))
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
			Name:    common.RunCompletionEventName,
			Payload: jsonPayload,
		}); err != nil {
			es.Logger.Error(err, "failed to send event")
			m.Nack()
			return
		}

		m.Ack()
	})

	if err != nil {
		es.Logger.Error(err, "failed to read from pubsub")
		return err
	}

	return nil
}

func (es *VaiEventingServer) runCompletionEventDataForRun(ctx context.Context, runId string) *common.RunCompletionEventData {
	job, err := es.PipelineJobClient.GetPipelineJob(ctx, &aiplatformpb.GetPipelineJobRequest{
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

func artifactsFilterData(job *aiplatformpb.PipelineJob) (componentCompletions []common.PipelineComponent) {
	//for _, artifactDef := range artifactDefs {
	// all move to kfp operator webhook
	//var evaluator *bexpr.Evaluator
	//var err error
	//
	//if artifactDef.Path.Filter != "" {
	//	evaluator, err = bexpr.CreateEvaluator(artifactDef.Path.Filter)
	//	if err != nil {
	//		continue
	//	}
	//}

	for _, task := range job.GetJobDetail().GetTaskDetails() {
		componentArtifactDetails := make([]common.ComponentArtifactDetails, 0, len(task.GetOutputs()))

		//if task.TaskName != artifactDef.Path.Locator.Name {
		//	continue
		//}

		for outputName, output := range task.GetOutputs() {
			//if outputName != artifactDef.Path.Locator.Artifact {
			//	continue
			//}
			//
			//if artifactDef.Path.Locator.Index >= len(output.Artifacts) {
			//	continue
			//}

			//artifact := output.Artifacts[artifactDef.Path.Locator.Index]
			//
			//if artifact.Uri == "" {
			//	continue
			//}
			artifacts := make([]common.ComponentOutputArtifact, 0)
			for _, artifact := range output.GetArtifacts() {
				metadata := artifact.Metadata.AsMap()
				artifacts = append(artifacts, common.ComponentOutputArtifact{
					Uri:      artifact.Uri,
					Metadata: metadata,
				})
			}

			//if evaluator != nil {
			//	matched, err := evaluator.Evaluate(artifact.Metadata.AsMap())
			//	// evaluator errors on missing properties
			//	if err != nil {
			//		continue
			//	}
			//	if !matched {
			//		continue
			//	}
			//}

			componentArtifactDetails = append(componentArtifactDetails, common.ComponentArtifactDetails{ArtifactName: outputName, Artifacts: artifacts})
		}
		componentCompletions = append(componentCompletions, common.PipelineComponent{Name: task.TaskName, ComponentArtifactDetails: componentArtifactDetails})
	}
	//}

	return componentCompletions
}

func (es *VaiEventingServer) toRunCompletionEventData(job *aiplatformpb.PipelineJob, runId string) *common.RunCompletionEventData {
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
		ComponentCompletion:   artifactsFilterData(job),
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
