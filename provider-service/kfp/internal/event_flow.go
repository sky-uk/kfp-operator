package internal

import (
	"context"
	"encoding/json"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
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
			rce, err := kef.toRunCompletionEventData(message)
			if err != nil {
				kef.errorOut <- err
			} else {
				kef.out <- rce
			}
		}
	}()
	return kef
}

func (kef *KfpEventFlow) To(s Sink[StreamMessage[*common.RunCompletionEventData]]) {
	for message := range kef.out {
		s.In() <- message
	}
}

func (kef *KfpEventFlow) Error(s Sink[error]) {
	for errorMessage := range kef.errorOut {
		s.In() <- errorMessage
	}
}

func NewKfpEventFlow(ctx context.Context, config KfpProviderConfig) (*KfpEventFlow, error) {
	logger := common.LoggerFromContext(ctx)

	kfpApi, err := ConnectToKfpApi(config.Parameters.GrpcKfpApiAddress)
	if err != nil {
		logger.Error(err, "failed to connect to Kubeflow API", "address", config.Parameters.GrpcKfpApiAddress)
		return nil, err
	}

	metadataStore, err := ConnectToMetadataStore(config.Parameters.GrpcMetadataStoreAddress)
	if err != nil {
		logger.Error(err, "failed to connect to metadata store", "address", config.Parameters.GrpcMetadataStoreAddress)
		return nil, err
	}

	return &KfpEventFlow{
		ProviderConfig: config,
		MetadataStore:  metadataStore,
		KfpApi:         kfpApi,
		Logger:         logger,
		context:        ctx,
		in:             make(chan StreamMessage[*unstructured.Unstructured]),
		out:            make(chan StreamMessage[*common.RunCompletionEventData]),
		errorOut:       make(chan error),
	}, nil
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
