package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/flow"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/ml_metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"os"
	"strconv"
	"strings"
)

var (
	argoWorkflowsGvr = schema.GroupVersionResource{
		Group:    workflow.Group,
		Version:  workflow.Version,
		Resource: workflow.WorkflowPlural,
	}
)

const (
	pipelineRunIdLabel           = "pipeline/runid"
	workflowPhaseLabel           = "workflows.argoproj.io/phase"
	workflowUpdateTriggeredLabel = "pipelines.kubeflow.org/events-published"
	pipelineSpecAnnotationName   = "pipelines.kubeflow.org/pipeline_spec"
)

type KfpEventSourceConfig struct {
	KfpNamespace string `yaml:"kfpNamespace"`
}

type KfpEventSource struct {
	K8sClient                        K8sClient
	RunCompletionEventConversionFlow streams.Flow
	Logger                           logr.Logger
	out                              chan any
}

type KfpEventFlow struct {
	ProviderConfig KfpProviderConfig
	MetadataStore  MetadataStore
	KfpApi         KfpApi
	Logger         logr.Logger
	context        context.Context
}

type PipelineSpec struct {
	Name string `json:"name"`
}

func NewKfpEventSource(ctx context.Context, provider string, namespace string) (*KfpEventSource, error) {
	logger := common.LoggerFromContext(ctx)
	k8sClient, err := NewK8sClient()
	if err != nil {
		logger.Error(err, "failed to initialise K8s Client")

		return nil, err
	}

	config := &KfpProviderConfig{
		Name: provider,
	}

	if err = LoadProvider(ctx, k8sClient.Client, provider, namespace, config); err != nil {
		logger.Error(err, "failed to load provider", "name", provider, "namespace", namespace)
		return nil, err
	}

	metadataStore, err := ConnectToMetadataStore(config.Parameters.GrpcMetadataStoreAddress)
	if err != nil {
		logger.Error(err, "failed to connect to metadata store", "address", config.Parameters.GrpcMetadataStoreAddress)
		return nil, err
	}

	kfpApi, err := ConnectToKfpApi(config.Parameters.GrpcKfpApiAddress)
	if err != nil {
		logger.Error(err, "failed to connect to Kubeflow API", "address", config.Parameters.GrpcKfpApiAddress)
		return nil, err
	}

	eventFlow := KfpEventFlow{
		ProviderConfig: *config,
		MetadataStore:  metadataStore,
		KfpApi:         kfpApi,
		Logger:         logger,
		context:        ctx,
	}

	kfpEventDataSource := &KfpEventSource{
		K8sClient:                        *k8sClient,
		RunCompletionEventConversionFlow: eventFlow.ToRCE(),
		Logger:                           logger,
		out:                              make(chan any),
	}

	go func() {
		if err := kfpEventDataSource.start(ctx, config.Parameters.KfpNamespace); err != nil {
			logger.Error(err, "failed to start KFP event source")
			os.Exit(1)
		}
	}()

	return kfpEventDataSource, nil
}

func (es *KfpEventSource) Via(operator streams.Flow) streams.Flow {
	flow.DoStream(es, operator)
	return operator
}

func (ef *KfpEventFlow) ToRCE() streams.Flow {
	// filterEmptyMessages := flow.NewFilter(func(data StreamMessage[*common.RunCompletionEventData]) bool {
	// 	fmt.Println(data.Message)
	// 	return data.Message != nil
	// }, 1)
	// 	fmt.Println("TORCE")
	return flow.NewMap(ef.toRunCompletionEventData, 1)
}

func (ef *KfpEventFlow) toRunCompletionEventData(message StreamMessage[*unstructured.Unstructured]) StreamMessage[*common.RunCompletionEventData] {
	runCompletionEventData, err := ef.eventForWorkflow(ef.context, message.Message)
	if err != nil {
		message.OnFailureHandler()
		return StreamMessage[*common.RunCompletionEventData]{}
	}
	return StreamMessage[*common.RunCompletionEventData]{
		Message:            runCompletionEventData,
		OnCompleteHandlers: message.OnCompleteHandlers,
	}
}

func (es *KfpEventSource) Out() <-chan any {
	return es.out
}

func (es *KfpEventSource) start(ctx context.Context, kfpNamespace string) error {
	es.Logger.Info("starting KFP event data source...")
	kfpSdkVersionExistsRequirement, err := labels.NewRequirement(
		pipelineRunIdLabel,
		selection.Exists,
		[]string{},
	)
	if err != nil {
		es.Logger.Error(err, "failed to construct requirement")
		return err
	}
	workflowUpdateTriggeredRequirement, err := labels.NewRequirement(
		workflowUpdateTriggeredLabel,
		selection.NotEquals,
		[]string{strconv.FormatBool(true)},
	)
	if err != nil {
		es.Logger.Error(err, "failed to construct requirement")
		return err
	}

	kfpWorkflowListOptions := metav1.ListOptions{
		LabelSelector: labels.NewSelector().
			Add(*kfpSdkVersionExistsRequirement).
			Add(*workflowUpdateTriggeredRequirement).
			String(),
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		es.K8sClient.Client,
		0,
		kfpNamespace,
		func(listOptions *metav1.ListOptions) {
			*listOptions = kfpWorkflowListOptions
		},
	)
	informer := factory.ForResource(argoWorkflowsGvr)
	sharedInformer := informer.Informer()

	handlerFuncs := cache.ResourceEventHandlerFuncs{}
	handlerFuncs.UpdateFunc = func(oldObj, newObj interface{}) {
		wf := newObj.(*unstructured.Unstructured)
		es.Logger.Info("sending run completion event data", "event", wf)

		select {
		case es.out <- StreamMessage[*unstructured.Unstructured]{
			Message: wf,
			OnCompleteHandlers: OnCompleteHandlers{
				OnSuccessHandler: func() {
					path := jsonPatchPath("metadata", "labels", workflowUpdateTriggeredLabel)
					patchPayload := fmt.Sprintf(`[{ "op": "replace", "path": "%s", "value": "true" }]`, path)
					_, err = es.
						K8sClient.
						Client.
						Resource(argoWorkflowsGvr).
						Namespace(wf.GetNamespace()).
						Patch(ctx, wf.GetName(), types.JSONPatchType, []byte(patchPayload), metav1.PatchOptions{})
					if err != nil {
						es.Logger.Error(err, "failed to patch resource")
						return
					}
				},
				OnFailureHandler: func() {},
			},
		}:
		case <-ctx.Done():
			es.Logger.Info("stopped reading from KFP event data source")
			return
		}

	}

	sharedInformer.AddEventHandler(handlerFuncs)
	sharedInformer.Run(ctx.Done())

	return nil
}

func jsonPatchPath(segments ...string) string {
	var builder strings.Builder

	for _, segment := range segments {
		tildeReplaced := strings.Replace(segment, "~", "~0", -1)
		slashReplaced := strings.Replace(tildeReplaced, "/", "~1", -1)
		builder.WriteString("/")
		builder.WriteString(slashReplaced)
	}

	return builder.String()
}

func (ef *KfpEventFlow) eventForWorkflow(ctx context.Context, workflow *unstructured.Unstructured) (*common.RunCompletionEventData, error) {
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
