package event

import (
	"context"
	"encoding/json"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type EventFlow struct {
	ProviderConfig config.KfpProviderConfig
	MetadataStore  client.MetadataStore
	KfpApi         client.KfpApi
	Logger         logr.Logger
	context        context.Context
	in             chan StreamMessage[*unstructured.Unstructured]
	out            chan StreamMessage[*common.RunCompletionEventData]
	errorOut       chan error
}

const (
	pipelineRunIdLabel           = "pipeline/runid"
	workflowPhaseLabel           = "workflows.argoproj.io/phase"
	workflowUpdateTriggeredLabel = "pipelines.kubeflow.org/events-published"
	pipelineSpecAnnotationName   = "pipelines.kubeflow.org/pipeline_spec"
)

type PipelineSpec struct {
	Name string `json:"name"`
}

func (ef *EventFlow) In() chan<- StreamMessage[*unstructured.Unstructured] {
	return ef.in
}

func (ef *EventFlow) Out() <-chan StreamMessage[*common.RunCompletionEventData] {
	return ef.out
}

func (ef *EventFlow) ErrOut() <-chan error {
	return ef.errorOut
}

func (ef *EventFlow) From(outlet Outlet[StreamMessage[*unstructured.Unstructured]]) Flow[StreamMessage[*unstructured.Unstructured], StreamMessage[*common.RunCompletionEventData], error] {
	go func() {
		for message := range outlet.Out() {
			ef.In() <- message
		}
	}()
	return ef
}

func (ef *EventFlow) To(inlet Inlet[StreamMessage[*common.RunCompletionEventData]]) {
	go func() {
		for message := range ef.out {
			inlet.In() <- message
		}
	}()
}

func (ef *EventFlow) Error(inlet Inlet[error]) {
	for errorMessage := range ef.errorOut {
		inlet.In() <- errorMessage
	}
}

func NewEventFlow(ctx context.Context, config config.KfpProviderConfig, kfpApi client.KfpApi, metadataStore client.MetadataStore) (*EventFlow, error) {
	logger := common.LoggerFromContext(ctx)

	flow := &EventFlow{
		ProviderConfig: config,
		MetadataStore:  metadataStore,
		KfpApi:         kfpApi,
		Logger:         logger,
		context:        ctx,
		in:             make(chan StreamMessage[*unstructured.Unstructured]),
		out:            make(chan StreamMessage[*common.RunCompletionEventData]),
		errorOut:       make(chan error),
	}

	go flow.subscribeAndConvert()

	return flow, nil
}

func (ef *EventFlow) subscribeAndConvert() {
	for msg := range ef.in {
		runCompletionEvent, err := ef.toRunCompletionEventData(msg)
		if err != nil {
			msg.OnFailureHandler()
			ef.errorOut <- err
		} else {
			ef.out <- runCompletionEvent
		}
	}
}

func (ef *EventFlow) toRunCompletionEventData(message StreamMessage[*unstructured.Unstructured]) (StreamMessage[*common.RunCompletionEventData], error) {
	runCompletionEventData, err := ef.eventForWorkflow(ef.context, message.Message)
	if err != nil {
		message.OnFailureHandler()
		return StreamMessage[*common.RunCompletionEventData]{}, err
	}
	return StreamMessage[*common.RunCompletionEventData]{
		Message:            runCompletionEventData,
		OnCompleteHandlers: message.OnCompleteHandlers,
	}, nil
}

func (ef *EventFlow) eventForWorkflow(ctx context.Context, workflow *unstructured.Unstructured) (*common.RunCompletionEventData, error) {
	status, hasFinished := runCompletionStatus(workflow)
	if !hasFinished {
		ef.Logger.V(2).Info("ignoring workflow that hasn't finished yet")
		return nil, nil
	}

	workflowName := workflow.GetName()

	modelArtifacts, err := ef.MetadataStore.GetServingModelArtifact(ctx, workflowName)
	if err != nil {
		ef.Logger.Error(err, "failed to retrieve serving model artifact")
		return nil, err
	}

	runId := workflow.GetLabels()[pipelineRunIdLabel]
	resourceReferences, err := ef.KfpApi.GetResourceReferences(ctx, runId)
	if err != nil {
		ef.Logger.Error(err, "failed to retrieve resource references")
		return nil, err
	}

	// For compatability with resources created with v0.3.0 and older
	if resourceReferences.PipelineName.Empty() {
		pipelineName := getPipelineName(workflow)
		if pipelineName == "" {
			ef.Logger.Info("failed to get pipeline name from workflow")
			return nil, nil
		}

		resourceReferences.PipelineName.Name = pipelineName
	}

	return &common.RunCompletionEventData{
		Status:                status,
		PipelineName:          resourceReferences.PipelineName,
		RunConfigurationName:  resourceReferences.RunConfigurationName.NonEmptyPtr(),
		RunName:               resourceReferences.RunName.NonEmptyPtr(),
		RunId:                 runId,
		ServingModelArtifacts: modelArtifacts,
		PipelineComponents:    nil,
		Provider:              ef.ProviderConfig.Name,
	}, nil
}

func runCompletionStatus(workflow *unstructured.Unstructured) (common.RunCompletionStatus, bool) {
	switch workflow.GetLabels()[workflowPhaseLabel] {
	case string(argo.WorkflowSucceeded):
		return common.RunCompletionStatuses.Succeeded, true
	case string(argo.WorkflowFailed), string(argo.WorkflowError):
		return common.RunCompletionStatuses.Failed, true
	default:
		return "", false
	}
}

func getPipelineNameFromAnnotation(workflow *unstructured.Unstructured) string {
	specString := workflow.GetAnnotations()[pipelineSpecAnnotationName]
	spec := &PipelineSpec{}
	if err := json.Unmarshal([]byte(specString), spec); err != nil {
		return ""
	}

	return spec.Name
}

func getPipelineName(workflow *unstructured.Unstructured) (name string) {
	if name = getPipelineNameFromAnnotation(workflow); name == "" {
		name = getPipelineNameFromEntrypoint(workflow)
	}

	return name
}

func getPipelineNameFromEntrypoint(workflow *unstructured.Unstructured) string {
	name, ok, err := unstructured.NestedString(workflow.Object, "spec", "entrypoint")

	if !ok || err != nil {
		return ""
	}

	return name
}
