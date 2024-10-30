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
	ProviderConfig KfpProviderConfig
	K8sClient      K8sClient
	MetadataStore  MetadataStore
	KfpApi         KfpApi
	Logger         logr.Logger
	out            chan any
}

type PipelineSpec struct {
	Name string `json:"name"`
}

func NewKfpEventSource(ctx context.Context, provider string, namespace string) (*KfpEventSource, error) {
	logger := common.LoggerFromContext(ctx)
	k8sClient, err := NewK8sClient()
	if err != nil {
		return nil, err
	}

	config := &KfpProviderConfig{
		Name: provider,
	}

	if err = LoadProvider[KfpProviderConfig](ctx, k8sClient.Client, provider, namespace, config); err != nil {
		return nil, err
	}

	metadataStore, err := ConnectToMetadataStore(config.Parameters.GrpcMetadataStoreAddress)
	if err != nil {
		return nil, err
	}

	kfpApi, err := ConnectToKfpApi(config.Parameters.GrpcKfpApiAddress)
	if err != nil {
		return nil, err
	}

	kfpEventDataSource := &KfpEventSource{
		K8sClient:      *k8sClient,
		ProviderConfig: *config,
		MetadataStore:  metadataStore,
		KfpApi:         kfpApi,
		Logger:         logger,
		out:            make(chan any),
	}

	go func() {
		err := kfpEventDataSource.start(ctx, config.Parameters.KfpNamespace)
		if err != nil {
			logger.Error(err, "Failed to start KFP event source")
			os.Exit(1)
		}
	}()

	return kfpEventDataSource, nil
}

func (es *KfpEventSource) Via(operator streams.Flow) streams.Flow {
	flow.DoStream(es, operator)
	return operator
}

func (es *KfpEventSource) Out() <-chan any {
	return es.out
}

func (es *KfpEventSource) start(ctx context.Context, kfpNamespace string) error {
	es.Logger.Info("starting KFP event data source...")
	kfpSdkVersionExistsRequirement, err := labels.NewRequirement(pipelineRunIdLabel, selection.Exists, []string{})
	if err != nil {
		es.Logger.Error(err, "failed to construct requirement")
		return err
	}
	workflowUpdateTriggeredRequirement, err := labels.NewRequirement(workflowUpdateTriggeredLabel, selection.NotEquals, []string{strconv.FormatBool(true)})
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

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(es.K8sClient.Client, 0, kfpNamespace, func(listOptions *metav1.ListOptions) {
		*listOptions = kfpWorkflowListOptions
	})

	ctx, cancel := context.WithCancel(ctx)

	informer := factory.ForResource(argoWorkflowsGvr)

	sharedInformer := informer.Informer()
	handlerFuncs := cache.ResourceEventHandlerFuncs{}
	handlerFuncs.UpdateFunc = func(oldObj, newObj interface{}) {
		es.Logger.Info("detected update event")

		wf := newObj.(*unstructured.Unstructured)

		runCompletionEventData, err := es.eventForWorkflow(ctx, wf)

		if err != nil {
			cancel() //force client to disconnect
			return
		}

		if runCompletionEventData == nil {
			return
		}

		es.Logger.Info("sending run completion event data", "event", runCompletionEventData)
		select {
		case es.out <- StreamMessage{
			RunCompletionEventData: *runCompletionEventData,
			OnCompleteHandlers: OnCompleteHandlers{
				OnSuccessHandler: func() {
					path := jsonPatchPath("metadata", "labels", workflowUpdateTriggeredLabel)
					patchPayload := fmt.Sprintf(`[{ "op": "replace", "path": "%s", "value": "true" }]`, path)
					_, err = es.K8sClient.Client.Resource(argoWorkflowsGvr).Namespace(wf.GetNamespace()).Patch(ctx, wf.GetName(), types.JSONPatchType, []byte(patchPayload), metav1.PatchOptions{})
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

func (es *KfpEventSource) eventForWorkflow(ctx context.Context, workflow *unstructured.Unstructured) (*common.RunCompletionEventData, error) {
	status, hasFinished := runCompletionStatus(workflow)
	if !hasFinished {
		es.Logger.V(2).Info("ignoring workflow that hasn't finished yet")
		return nil, nil
	}

	workflowName := workflow.GetName()

	modelArtifacts, err := es.MetadataStore.GetServingModelArtifact(ctx, workflowName)
	if err != nil {
		es.Logger.Error(err, "failed to retrieve serving model artifact")
		return nil, err
	}

	runId := workflow.GetLabels()[pipelineRunIdLabel]
	resourceReferences, err := es.KfpApi.GetResourceReferences(ctx, runId)
	if err != nil {
		es.Logger.Error(err, "failed to retrieve resource references")
		return nil, err
	}

	// For compatability with resources created with v0.3.0 and older
	if resourceReferences.PipelineName.Empty() {
		pipelineName := getPipelineName(workflow)
		if pipelineName == "" {
			es.Logger.Info("failed to get pipeline name from workflow")
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
		Provider:              es.ProviderConfig.Name,
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
