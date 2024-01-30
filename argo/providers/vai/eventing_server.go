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
	"github.com/hashicorp/go-bexpr"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
		println(fmt.Sprintf("LOG ENTRY=%+v", logEntry))
		println(fmt.Sprintf("DATA=%s", string(m.Data)))

		gvr, namespacedName, err := gvrAndNamespacedNameForRunLabels(logEntry.Labels)
		if err != nil {
			es.Logger.Error(err, "failed to extract GVR from labels")
			m.Nack()
			return
		}

		artifactDefs, err := es.GetRunArtifactDefinitions(ctx, namespacedName, gvr)
		if err != nil {
			es.Logger.Error(err, fmt.Sprintf("failed to lookup run/runconfiguration %s", gvr.Resource))
			m.Nack()
			return
		}

		pipelineJobId, ok := logEntry.Resource.Labels["pipeline_job_id"]
		if !ok {
			es.Logger.Error(err, fmt.Sprintf("logEntry did not contain pipeline_job_id %+v", logEntry))
			m.Nack()
			return
		}
		run := VAIRun{
			Labels:    logEntry.Labels,
			RunId:     pipelineJobId,
			Artifacts: artifactDefs,
		}

		event := es.runCompletionEventForRun(stream.Context(), run)
		if event == nil {
			es.Logger.Error(err, fmt.Sprintf("failed to convert to run completion event %+v", run))
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
		return err
	}

	return nil
}

func (es *VaiEventingServer) runCompletionEventForRun(ctx context.Context, run VAIRun) *common.RunCompletionEvent {
	job, err := es.PipelineJobClient.GetPipelineJob(ctx, &aiplatformpb.GetPipelineJobRequest{
		Name: es.ProviderConfig.pipelineJobName(run.RunId),
	})
	if err != nil {
		es.Logger.Error(err, "could not fetch pipeline job")
		return nil
	}
	if job == nil {
		es.Logger.Error(nil, "expected pipeline job not found", "run-id", run.RunId)
		return nil
	}

	return es.toRunCompletionEvent(job, run)
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

func artifactsForJob(job *aiplatformpb.PipelineJob, artifactDefs []pipelinesv1.OutputArtifact) (artifacts []common.Artifact) {
	for _, artifactDef := range artifactDefs {
		var evaluator *bexpr.Evaluator
		var err error

		if artifactDef.Path.Filter != "" {
			evaluator, err = bexpr.CreateEvaluator(artifactDef.Path.Filter)
			if err != nil {
				continue
			}
		}

		for _, task := range job.GetJobDetail().GetTaskDetails() {
			if task.TaskName != artifactDef.Path.Locator.Component {
				continue
			}

			for outputName, output := range task.GetOutputs() {
				if outputName != artifactDef.Path.Locator.Artifact {
					continue
				}

				if artifactDef.Path.Locator.Index >= len(output.Artifacts) {
					continue
				}

				artifact := output.Artifacts[artifactDef.Path.Locator.Index]

				if artifact.Uri == "" {
					continue
				}

				if evaluator != nil {
					matched, err := evaluator.Evaluate(artifact.Metadata.AsMap())
					// evaluator errors on missing properties
					if err != nil {
						continue
					}
					if !matched {
						continue
					}
				}

				artifacts = append(artifacts, common.Artifact{Name: artifactDef.Name, Location: artifact.Uri})
			}
		}
	}

	return artifacts
}

func gvrAndNamespacedNameForRunLabels(runLabels map[string]string) (schema.GroupVersionResource, types.NamespacedName, error) {
	runConfigurationName, ok := runLabels[labels.RunConfigurationName]
	if ok {
		return base.RunConfigurationGVR, types.NamespacedName{Name: runConfigurationName, Namespace: runLabels[labels.RunConfigurationNamespace]}, nil
	}

	runName, ok := runLabels[labels.RunName]
	if ok {
		return base.RunGVR, types.NamespacedName{Name: runName, Namespace: runLabels[labels.RunNamespace]}, nil
	}

	return schema.GroupVersionResource{}, types.NamespacedName{}, fmt.Errorf("neither %s or %s provided in labels", labels.RunConfigurationName, labels.RunName)
}

func (es *VaiEventingServer) toRunCompletionEvent(job *aiplatformpb.PipelineJob, run VAIRun) *common.RunCompletionEvent {
	runCompletionStatus, completed := runCompletionStatus(job)

	if !completed {
		es.Logger.Error(nil, "expected pipeline job to have finished", "run-id", run.RunId)
		return nil
	}

	var runName, runConfigurationName, pipelineName common.NamespacedName

	pipelineName.Name = job.Labels[labels.PipelineName]
	if pipelineNamespace, ok := job.Labels[labels.PipelineNamespace]; ok {
		pipelineName.Namespace = pipelineNamespace
	}

	if legacyNamespace, ok := job.Labels[labels.LegacyNamespace]; ok {
		// For compatability with resources created with v0.3.0 and older
		runName = common.NamespacedName{
			Name:      run.RunId,
			Namespace: legacyNamespace,
		}
	} else {
		runName = common.NamespacedName{
			Name:      job.Labels[labels.RunName],
			Namespace: job.Labels[labels.RunNamespace]}
	}

	if legacyRunConfiguration, ok := job.Labels[labels.LegacyRunConfiguration]; ok {
		// For compatability with resources created with v0.3.0 and older
		runConfigurationName = common.NamespacedName{
			Name: legacyRunConfiguration,
		}
	} else {
		runConfigurationName = common.NamespacedName{
			Name:      job.Labels[labels.RunConfigurationName],
			Namespace: job.Labels[labels.RunConfigurationNamespace]}
	}

	return &common.RunCompletionEvent{
		Status:                runCompletionStatus,
		PipelineName:          pipelineName,
		RunConfigurationName:  runConfigurationName.NonEmptyPtr(),
		RunName:               runName.NonEmptyPtr(),
		RunId:                 run.RunId,
		Artifacts:             artifactsForJob(job, run.Artifacts),
		ServingModelArtifacts: modelServingArtifactsForJob(job),
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
