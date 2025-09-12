package sources

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/internal/log"
	"strconv"
	"strings"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	"github.com/go-logr/logr"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
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
	workflowUpdateTriggeredLabel = "pipelines.kubeflow.org/events-published"
	disableCacheResync           = 0
)

type WorkflowSource struct {
	K8sClient K8sClient
	Logger    logr.Logger
	out       chan StreamMessage[*unstructured.Unstructured]
}

func (ws *WorkflowSource) Out() <-chan StreamMessage[*unstructured.Unstructured] {
	return ws.out
}

func NewWorkflowSource(ctx context.Context, namespace string, k8sClient K8sClient) (*WorkflowSource, error) {
	logger := log.LoggerFromContext(ctx)

	workflowSource := &WorkflowSource{
		K8sClient: k8sClient,
		Logger:    logger,
		out:       make(chan StreamMessage[*unstructured.Unstructured]),
	}

	go func() {
		if err := workflowSource.start(ctx, namespace); err != nil {
			panic(err)
		}
	}()

	return workflowSource, nil
}

func (ws *WorkflowSource) start(ctx context.Context, namespace string) error {
	ws.Logger.Info("starting workflow event data source...")
	kfpOperatorExistsRequirement, err := labels.NewRequirement(
		pipelineRunIdLabel,
		selection.Exists,
		[]string{},
	)
	if err != nil {
		ws.Logger.Error(err, "failed to construct requirement")
		return err
	}
	workflowUpdateTriggeredRequirement, err := labels.NewRequirement(
		workflowUpdateTriggeredLabel,
		selection.NotEquals,
		[]string{strconv.FormatBool(true)},
	)
	if err != nil {
		ws.Logger.Error(err, "failed to construct requirement")
		return err
	}

	workflowListOptions := metav1.ListOptions{
		LabelSelector: labels.NewSelector().
			Add(*kfpOperatorExistsRequirement).
			Add(*workflowUpdateTriggeredRequirement).
			String(),
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		ws.K8sClient.Client,
		disableCacheResync,
		namespace,
		func(listOptions *metav1.ListOptions) {
			*listOptions = workflowListOptions
		},
	)
	informer := factory.ForResource(argoWorkflowsGvr)
	sharedInformer := informer.Informer()

	handlerFuncs := cache.ResourceEventHandlerFuncs{}
	handlerFuncs.UpdateFunc = func(oldObj, newObj interface{}) {
		wf := newObj.(*unstructured.Unstructured)
		ws.Logger.Info("received workflow update event", "workflow", wf.GetName())

		select {
		case ws.out <- StreamMessage[*unstructured.Unstructured]{
			Message: wf,
			OnCompleteHandlers: OnCompleteHandlers{
				OnSuccessHandler: func() {
					path := jsonPatchPath("metadata", "labels", workflowUpdateTriggeredLabel)
					patchPayload := fmt.Sprintf(`[{ "op": "replace", "path": "%s", "value": "true" }]`, path)
					_, err = ws.
						K8sClient.
						Client.
						Resource(argoWorkflowsGvr).
						Namespace(wf.GetNamespace()).
						Patch(ctx, wf.GetName(), types.JSONPatchType, []byte(patchPayload), metav1.PatchOptions{})
					if err != nil {
						ws.Logger.Error(err, "failed to patch resource")
						return
					}
				},
				OnRecoverableFailureHandler: func() {},
			},
		}:
		case <-ctx.Done():
			ws.Logger.Info("stopped reading from workflow data source")
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
