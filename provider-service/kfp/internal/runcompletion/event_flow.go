package runcompletion

import (
	"context"
	"encoding/json"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/pkg/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type kfpApi interface {
	GetResourceReferences(
		ctx context.Context,
		runId string,
	) (resource.References, error)
}

type metadataStore interface {
	GetArtifactsForRun(
		ctx context.Context,
		runId string,
	) ([]common.PipelineComponent, error)
}

type EventFlow struct {
	ProviderConfig config.Config
	metadataStore  metadataStore
	kfpApi         kfpApi
	Logger         logr.Logger
	in             chan StreamMessage[*unstructured.Unstructured]
	out            chan StreamMessage[*common.RunCompletionEventData]
	errorOut       chan error
}

const (
	pipelineRunIdLabel         = "pipeline/runid"
	workflowPhaseLabel         = "workflows.argoproj.io/phase"
	pipelineSpecAnnotationName = "pipelines.kubeflow.org/pipeline_spec"
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

func NewEventFlow(
	ctx context.Context,
	config config.Config,
	kfpApiClient kfpApi,
	metadataStore metadataStore,
) (*EventFlow, error) {
	logger := logr.FromContextOrDiscard(ctx)

	flow := &EventFlow{
		ProviderConfig: config,
		metadataStore:  metadataStore,
		kfpApi:         kfpApiClient,
		Logger:         logger,
		in:             make(chan StreamMessage[*unstructured.Unstructured]),
		out:            make(chan StreamMessage[*common.RunCompletionEventData]),
		errorOut:       make(chan error),
	}

	go flow.subscribeAndConvert(ctx)

	return flow, nil
}

func (ef *EventFlow) subscribeAndConvert(ctx context.Context) {
	for msg := range ef.in {
		runCompletionEvent, err := ef.toRunCompletionEventData(ctx, msg)
		if err != nil {
			msg.OnRecoverableFailureHandler()
			ef.errorOut <- err
		} else {
			ef.out <- runCompletionEvent
		}
	}
}

func (ef *EventFlow) toRunCompletionEventData(ctx context.Context, message StreamMessage[*unstructured.Unstructured]) (StreamMessage[*common.RunCompletionEventData], error) {
	runCompletionEventData, err := ef.eventForWorkflow(ctx, message.Message)
	if err != nil {
		message.OnRecoverableFailureHandler()
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
	if workflowName == "" {
		return nil, nil
	}

	runId := workflow.GetLabels()[pipelineRunIdLabel]
	resourceReferences, err := ef.kfpApi.GetResourceReferences(ctx, runId)
	if err != nil {
		ef.Logger.Error(err, "failed to retrieve resource references")
		return nil, err
	}

	pipelineComponents, err := ef.metadataStore.GetArtifactsForRun(ctx, runId)
	if err != nil {
		ef.Logger.Error(err, "failed to retrieve pipeline components")
		return nil, err
	}

	return &common.RunCompletionEventData{
		Status:               status,
		PipelineName:         resourceReferences.PipelineName,
		RunConfigurationName: resourceReferences.RunConfigurationName.NonEmptyPtr(),
		RunName:              resourceReferences.RunName.NonEmptyPtr(),
		RunId:                runId,
		PipelineComponents:   pipelineComponents,
		Provider:             ef.ProviderConfig.ProviderName,
		RunStartTime:         resourceReferences.CreatedAt,
		RunEndTime:           resourceReferences.FinishedAt,
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
