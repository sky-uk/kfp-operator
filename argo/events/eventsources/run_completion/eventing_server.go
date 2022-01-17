package run_completion

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"pipelines.kubeflow.org/events/eventsources/generic"
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
	kfpSdkVersionLabel           = "pipelines.kubeflow.org/kfp_sdk_version"
	workflowPhaseLabel           = "workflows.argoproj.io/phase"
	workflowUpdateTriggeredLabel = "pipelines.kubeflow.org/events-published"
	pipelineSpecAnnotationName = "pipelines.kubeflow.org/pipeline_spec"
	runCompletionEventName     = "run-completion"
)

type EventSourceConfig struct {
	KfpNamespace string `yaml:"kfpNamespace"`
}

type EventingServer struct {
	generic.UnimplementedEventingServer
	K8sClient     dynamic.Interface
	MetadataStore MetadataStore
	Logger        logr.Logger
}

type ServingModelArtifact struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type RunCompletionStatus string

const (
	Succeeded  RunCompletionStatus = "succeeded"
	Failed  RunCompletionStatus = "failed"
)

type RunCompletionEvent struct {
	Status RunCompletionStatus
	PipelineName          string                 `json:"pipelineName"`
	ServingModelArtifacts []ServingModelArtifact `json:"servingModelArtifacts"`
}

type PipelineSpec struct {
	Name string `json:"name"`
}

func getPipelineName(workflow *unstructured.Unstructured) (string, error) {
	specString := workflow.GetAnnotations()[pipelineSpecAnnotationName]
	spec := &PipelineSpec{}
	if err := json.Unmarshal([]byte(specString), spec); err != nil {
		return "", err
	}

	if spec.Name == "" {
		return "", fmt.Errorf("workflow has empty pipeline name")
	}

	return spec.Name, nil
}

func (es *EventingServer) StartEventSource(source *generic.EventSource, stream generic.Eventing_StartEventSourceServer) error {
	eventsourceConfig := EventSourceConfig{}

	if err := yaml.Unmarshal(source.Config, &eventsourceConfig); err != nil {
		es.Logger.Error(err, "failed to parse event source configuration")
		return err
	}

	es.Logger.Info("starting stream", "eventsource", eventsourceConfig)

	kfpSdkVersionExistsRequirement, err := labels.NewRequirement(kfpSdkVersionLabel, selection.Exists, []string{})
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

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(es.K8sClient, 0, eventsourceConfig.KfpNamespace, func(listOptions *metav1.ListOptions) {
		*listOptions = kfpWorkflowListOptions
	})

	informer := factory.ForResource(argoWorkflowsGvr)

	sharedInformer := informer.Informer()
	handlerFuncs := cache.ResourceEventHandlerFuncs{}
	handlerFuncs.UpdateFunc = func(oldObj, newObj interface{}) {
		es.Logger.V(2).Info("detected update event")

		workflow := newObj.(*unstructured.Unstructured)

		if !workflowHasFinished(workflow) {
			es.Logger.V(2).Info("ignoring workflow that hasn't finished yet")
			return
		}

		workflowName := workflow.GetName()
		pipelineName, err := getPipelineName(workflow)

		if err != nil {
			es.Logger.Error(err, "failed to get pipeline name from workflow")
			return
		}

		var runCompletionEvent RunCompletionEvent

		if workflowHasSucceeded(workflow) {
			modelArtifacts, err := es.MetadataStore.GetServingModelArtifact(stream.Context(), workflowName)

			if err != nil {
				es.Logger.Error(err, "failed to retrieve serving model artifact")
				return
			}

			runCompletionEvent = RunCompletionEvent{
				Status: Succeeded,
				PipelineName:          pipelineName,
				ServingModelArtifacts: modelArtifacts,
			}
		} else {
			runCompletionEvent = RunCompletionEvent{
				Status: Failed,
				PipelineName: pipelineName,
			}
		}

		jsonPayload, err := json.Marshal(runCompletionEvent)

		if err != nil {
			es.Logger.Error(err, "failed to serialise event")
			return
		}

		es.Logger.V(1).Info("sending run completion event", "event", runCompletionEvent)
		if err = stream.Send(&generic.Event{
			Name:    runCompletionEventName,
			Payload: jsonPayload,
		}); err != nil {
			es.Logger.Error(err, "failed to send event")
			return
		}

		path := jsonPatchPath("metadata", "labels", workflowUpdateTriggeredLabel)
		patchPayload := fmt.Sprintf(`[{ "op": "replace", "path": "%s", "value": "true" }]`, path)
		_, err = es.K8sClient.Resource(argoWorkflowsGvr).Namespace(workflow.GetNamespace()).Patch(stream.Context(), workflow.GetName(), types.JSONPatchType, []byte(patchPayload), metav1.PatchOptions{})
		if err != nil {
			es.Logger.Error(err, "failed to patch resource")
			return
		}
	}

	sharedInformer.AddEventHandler(handlerFuncs)
	sharedInformer.Run(stream.Context().Done())

	return nil
}

func workflowHasSucceeded(workflow *unstructured.Unstructured) bool {
	return workflow.GetLabels()[workflowPhaseLabel] == string(argo.WorkflowSucceeded)
}

func workflowHasFinished(workflow *unstructured.Unstructured) bool {
	switch workflow.GetLabels()[workflowPhaseLabel] {
	case string(argo.WorkflowSucceeded), string(argo.WorkflowFailed), string(argo.WorkflowError):
		return true
	default:
		return false
	}
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
