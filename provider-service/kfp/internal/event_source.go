package internal

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	"github.com/go-logr/logr"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
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
	RunCompletionEventConversionFlow Flow[StreamMessage[*unstructured.Unstructured], StreamMessage[*common.RunCompletionEventData], error]
	Logger                           logr.Logger
	out                              chan StreamMessage[*unstructured.Unstructured]
}

type PipelineSpec struct {
	Name string `json:"name"`
}

func (kes *KfpEventSource) Out() <-chan StreamMessage[*unstructured.Unstructured] {
	return kes.out
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

	flow, err := NewKfpEventFlow(ctx, *config)
	if err != nil {
		logger.Error(err, "failed to create RunCompletionEvent conversion flow")
		return nil, err
	}

	kfpEventDataSource := &KfpEventSource{
		K8sClient:                        *k8sClient,
		Logger:                           logger,
		out:                              make(chan StreamMessage[*unstructured.Unstructured]),
		RunCompletionEventConversionFlow: flow,
	}

	go func() {
		if err := kfpEventDataSource.start(ctx, config.Parameters.KfpNamespace); err != nil {
			logger.Error(err, "failed to start KFP event source")
			os.Exit(1)
		}
	}()

	return kfpEventDataSource, nil
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
