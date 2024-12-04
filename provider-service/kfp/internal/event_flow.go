package internal

import (
	"context"
	"encoding/json"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/ml_metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type KfpEventFlow struct {
	ProviderConfig KfpProviderConfig
	MetadataStore  MetadataStore
	KfpApi         KfpApi
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

func (kef *KfpEventFlow) In() chan<- StreamMessage[*unstructured.Unstructured] {
	return kef.in
}

func (kef *KfpEventFlow) Out() <-chan StreamMessage[*common.RunCompletionEventData] {
	return kef.out
}

func (kef *KfpEventFlow) ErrOut() <-chan error {
	return kef.errorOut
}

func (kef *KfpEventFlow) From(outlet Outlet[StreamMessage[*unstructured.Unstructured]]) Flow[StreamMessage[*unstructured.Unstructured], StreamMessage[*common.RunCompletionEventData], error] {
	go func() {
		for message := range outlet.Out() {
			kef.In() <- message
		}
	}()
	return kef
}

func (kef *KfpEventFlow) To(inlet Inlet[StreamMessage[*common.RunCompletionEventData]]) {
	go func() {
		for message := range kef.out {
			inlet.In() <- message
		}
	}()
}

func (kef *KfpEventFlow) Error(inlet Inlet[error]) {
	for errorMessage := range kef.errorOut {
		inlet.In() <- errorMessage
	}
}

func CreateKfpApi(ctx context.Context, config KfpProviderConfig) (KfpApi, error) {
	logger := common.LoggerFromContext(ctx)
	kfpApi, err := ConnectToKfpApi(config.Parameters.GrpcKfpApiAddress)
	if err != nil {
		logger.Error(err, "failed to connect to Kubeflow API", "address", config.Parameters.GrpcKfpApiAddress)
		return nil, err
	}
	return kfpApi, nil
}

func CreateMetadataStore(ctx context.Context, config KfpProviderConfig) (MetadataStore, error) {
	logger := common.LoggerFromContext(ctx)
	metadataStore, err := ConnectToMetadataStore(config.Parameters.GrpcMetadataStoreAddress)
	if err != nil {
		logger.Error(err, "failed to connect to metadata store", "address", config.Parameters.GrpcMetadataStoreAddress)
		return nil, err
	}
	return metadataStore, nil
}

func NewKfpEventFlow(ctx context.Context, config KfpProviderConfig, kfpApi KfpApi, metadataStore MetadataStore) (*KfpEventFlow, error) {
	logger := common.LoggerFromContext(ctx)

	flow := &KfpEventFlow{
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

func (kef *KfpEventFlow) subscribeAndConvert() {
	for msg := range kef.in {
		runCompletionEvent, err := kef.toRunCompletionEventData(msg)
		if err != nil {
			msg.OnFailureHandler()
			kef.errorOut <- err
		} else {
			kef.out <- runCompletionEvent
		}
	}
}

func (kef *KfpEventFlow) toRunCompletionEventData(message StreamMessage[*unstructured.Unstructured]) (StreamMessage[*common.RunCompletionEventData], error) {
	runCompletionEventData, err := kef.eventForWorkflow(kef.context, message.Message)
	if err != nil {
		message.OnFailureHandler()
		return StreamMessage[*common.RunCompletionEventData]{}, err
	}
	return StreamMessage[*common.RunCompletionEventData]{
		Message:            runCompletionEventData,
		OnCompleteHandlers: message.OnCompleteHandlers,
	}, nil
}

func (kef *KfpEventFlow) eventForWorkflow(ctx context.Context, workflow *unstructured.Unstructured) (*common.RunCompletionEventData, error) {
	status, hasFinished := runCompletionStatus(workflow)
	if !hasFinished {
		kef.Logger.V(2).Info("ignoring workflow that hasn't finished yet")
		return nil, nil
	}

	workflowName := workflow.GetName()

	modelArtifacts, err := kef.MetadataStore.GetServingModelArtifact(ctx, workflowName)
	if err != nil {
		kef.Logger.Error(err, "failed to retrieve serving model artifact")
		return nil, err
	}

	runId := workflow.GetLabels()[pipelineRunIdLabel]
	resourceReferences, err := kef.KfpApi.GetResourceReferences(ctx, runId)
	if err != nil {
		kef.Logger.Error(err, "failed to retrieve resource references")
		return nil, err
	}

	// For compatability with resources created with v0.3.0 and older
	if resourceReferences.PipelineName.Empty() {
		pipelineName := getPipelineName(workflow)
		if pipelineName == "" {
			kef.Logger.Info("failed to get pipeline name from workflow")
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
		Provider:              kef.ProviderConfig.Name,
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
	name = getPipelineNameFromAnnotation(workflow)

	if name == "" {
		name = getPipelineNameFromEntrypoint(workflow)
	}

	return
}

func getPipelineNameFromEntrypoint(workflow *unstructured.Unstructured) string {
	name, ok, err := unstructured.NestedString(workflow.Object, "spec", "entrypoint")

	if !ok || err != nil {
		return ""
	}

	return name
}

func ConnectToMetadataStore(address string) (*GrpcMetadataStore, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcMetadataStore{
		MetadataStoreServiceClient: ml_metadata.NewMetadataStoreServiceClient(conn),
	}, nil
}

func ConnectToKfpApi(address string) (*GrpcKfpApi, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcKfpApi{
		RunServiceClient: go_client.NewRunServiceClient(conn),
		JobServiceClient: go_client.NewJobServiceClient(conn),
	}, nil
}
