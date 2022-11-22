package kfp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/base/generic"
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

type KfpEventingServer struct {
	generic.UnimplementedEventingServer
	K8sClient     dynamic.Interface
	MetadataStore MetadataStore
	KfpApi        KfpApi
	Logger        logr.Logger
}

type PipelineSpec struct {
	Name string `json:"name"`
}

func getPipelineNameFromAnnotation(workflow *unstructured.Unstructured) string {
	specString := workflow.GetAnnotations()[pipelineSpecAnnotationName]
	spec := &PipelineSpec{}
	if err := json.Unmarshal([]byte(specString), spec); err != nil {
		return ""
	}

	return spec.Name
}

func getPipelineNameFromEntrypoint(workflow *unstructured.Unstructured) string {
	name, ok, err := unstructured.NestedString(workflow.Object, "spec", "entrypoint")

	if !ok || err != nil {
		return ""
	}

	return name
}

func getPipelineName(workflow *unstructured.Unstructured) (name string) {
	name = getPipelineNameFromAnnotation(workflow)

	if name == "" {
		name = getPipelineNameFromEntrypoint(workflow)
	}

	return
}

func (es *KfpEventingServer) StartEventSource(source *generic.EventSource, stream generic.Eventing_StartEventSourceServer) error {
	eventsourceConfig := KfpEventSourceConfig{}

	if err := yaml.Unmarshal(source.Config, &eventsourceConfig); err != nil {
		es.Logger.Error(err, "failed to parse event source configuration")
		return err
	}

	es.Logger.Info("starting stream", "eventsource", eventsourceConfig)

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

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(es.K8sClient, 0, eventsourceConfig.KfpNamespace, func(listOptions *metav1.ListOptions) {
		*listOptions = kfpWorkflowListOptions
	})

	ctx, cancel := context.WithCancel(stream.Context())

	informer := factory.ForResource(argoWorkflowsGvr)

	sharedInformer := informer.Informer()
	handlerFuncs := cache.ResourceEventHandlerFuncs{}
	handlerFuncs.UpdateFunc = func(oldObj, newObj interface{}) {
		es.Logger.V(2).Info("detected update event")

		workflow := newObj.(*unstructured.Unstructured)

		runCompletionEvent, err := es.eventForWorkflow(ctx, workflow)

		if err != nil {
			cancel() //force client to disconnect
			return
		}

		if runCompletionEvent == nil {
			return
		}

		jsonPayload, err := json.Marshal(runCompletionEvent)

		if err != nil {
			es.Logger.Error(err, "failed to serialise event")
			return
		}

		es.Logger.V(1).Info("sending run completion event", "event", runCompletionEvent)
		if err = stream.Send(&generic.Event{
			Name:    RunCompletionEventName,
			Payload: jsonPayload,
		}); err != nil {
			es.Logger.Error(err, "failed to send event")
			return
		}

		path := jsonPatchPath("metadata", "labels", workflowUpdateTriggeredLabel)
		patchPayload := fmt.Sprintf(`[{ "op": "replace", "path": "%s", "value": "true" }]`, path)
		_, err = es.K8sClient.Resource(argoWorkflowsGvr).Namespace(workflow.GetNamespace()).Patch(ctx, workflow.GetName(), types.JSONPatchType, []byte(patchPayload), metav1.PatchOptions{})
		if err != nil {
			es.Logger.Error(err, "failed to patch resource")
			return
		}
	}

	sharedInformer.AddEventHandler(handlerFuncs)
	sharedInformer.Run(ctx.Done())

	return nil
}

func (es *KfpEventingServer) eventForWorkflow(ctx context.Context, workflow *unstructured.Unstructured) (*RunCompletionEvent, error) {
	status, hasFinished := runCompletionStatus(workflow)

	if !hasFinished {
		es.Logger.V(2).Info("ignoring workflow that hasn't finished yet")
		return nil, nil
	}

	workflowName := workflow.GetName()
	pipelineName := getPipelineName(workflow)

	if pipelineName == "" {
		es.Logger.Info("failed to get pipeline name from workflow")
		return nil, nil
	}

	modelArtifacts, err := es.MetadataStore.GetServingModelArtifact(ctx, workflowName)

	if err != nil {
		es.Logger.Error(err, "failed to retrieve serving model artifact")
		return nil, err
	}

	runId := workflow.GetLabels()[pipelineRunIdLabel]
	runConfigurationName, err := es.KfpApi.GetRunConfiguration(ctx, runId)

	if err != nil {
		es.Logger.Error(err, "failed to retrieve RunConfiguration name")
		return nil, err
	}

	return &RunCompletionEvent{
		Status:                status,
		PipelineName:          pipelineName,
		RunConfigurationName:  runConfigurationName,
		ServingModelArtifacts: modelArtifacts,
	}, nil
}

func runCompletionStatus(workflow *unstructured.Unstructured) (RunCompletionStatus, bool) {
	switch workflow.GetLabels()[workflowPhaseLabel] {
	case string(argo.WorkflowSucceeded):
		return Succeeded, true
	case string(argo.WorkflowFailed), string(argo.WorkflowError):
		return Failed, true
	default:
		return "", false
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
