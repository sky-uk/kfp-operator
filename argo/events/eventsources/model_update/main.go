package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
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
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net"
	"os"
	"path/filepath"
	"pipelines.kubeflow.org/events/eventsources/generic"
	"pipelines.kubeflow.org/events/logging"
	"pipelines.kubeflow.org/events/ml_metadata"
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
	pipelineSpecAnnotationName   = "pipelines.kubeflow.org/pipeline_spec"
	modelUpdateEventName 		 = "model-update"
)

type EventSourceConfig struct {
	KfpNamespace string `yaml:"kfpNamespace"`
}

type eventingServer struct {
	generic.UnimplementedEventingServer
	k8sClient dynamic.Interface
	metadataStore MetadataStore
	logger logr.Logger
}

type ModelUpdateEvent struct {
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

func (es *eventingServer) StartEventSource(source *generic.EventSource, stream generic.Eventing_StartEventSourceServer) error {
	es.logger.Info("starting stream", "eventsource", source)

	eventsourceConfig := EventSourceConfig{}

	if err := yaml.Unmarshal(source.Config, &eventsourceConfig); err != nil {
		es.logger.Error(err, "failed to parse event source configuration")
		return err
	}

	kfpSdkVersionExistsRequirement, err := labels.NewRequirement(kfpSdkVersionLabel, selection.Exists, []string{})
	if err != nil {
		es.logger.Error(err, "failed to construct requirement")
		return err
	}
	workflowUpdateTriggeredRequirement, err := labels.NewRequirement(workflowUpdateTriggeredLabel, selection.NotEquals, []string{strconv.FormatBool(true)})
	if err != nil {
		es.logger.Error(err, "failed to construct requirement")
		return err
	}

	kfpWorkflowListOptions := metav1.ListOptions{
		LabelSelector: labels.NewSelector().
			Add(*kfpSdkVersionExistsRequirement).
			Add(*workflowUpdateTriggeredRequirement).
			String(),
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(es.k8sClient, 0, eventsourceConfig.KfpNamespace, func(listOptions *metav1.ListOptions) {
		*listOptions = kfpWorkflowListOptions
	})

	informer := factory.ForResource(argoWorkflowsGvr)

	sharedInformer := informer.Informer()
	handlerFuncs := cache.ResourceEventHandlerFuncs{}
	handlerFuncs.UpdateFunc = func(oldObj, newObj interface{}) {
		es.logger.V(2).Info("detected update event")

		uNewObj := newObj.(*unstructured.Unstructured)

		if uNewObj.GetLabels()[workflowPhaseLabel] != string(argo.WorkflowSucceeded) {
			es.logger.V(2).Info("rejecting workflow that hasn't succeeded yet")
			return
		}

		workflowName := uNewObj.GetName()
		pipelineName, err := getPipelineName(uNewObj)

		if err != nil {
			es.logger.Error(err,"failed to get pipeline name from workflow")
			return
		}

		modelArtifacts, err := es.metadataStore.GetServingModelArtifact(stream.Context(), workflowName)

		if err != nil {
			es.logger.Error(err,"failed to retrieve workflow artifacts")
			return
		}

		jsonPayload, err := json.Marshal(ModelUpdateEvent{
			PipelineName:          pipelineName,
			ServingModelArtifacts: modelArtifacts,
		})

		if err != nil {
			es.logger.Error(err,"failed to serialise event")
			return
		}

		event := &generic.Event{
			Name:    modelUpdateEventName,
			Payload: jsonPayload,
		}

		es.logger.V(1).Info("sending event", "event", event)
		if err = stream.Send(event); err != nil {
			es.logger.Error(err, "failed to send event")
			return
		}

		path := jsonPatchPath("metadata", "labels", workflowUpdateTriggeredLabel)
		patchPayload := fmt.Sprintf(`[{ "op": "replace", "path": "%s", "value": "true" }]`, path)
		_, err = es.k8sClient.Resource(argoWorkflowsGvr).Namespace(uNewObj.GetNamespace()).Patch(stream.Context(), uNewObj.GetName(), types.JSONPatchType, []byte(patchPayload), metav1.PatchOptions{})
		if err != nil {
			es.logger.Error(err, "failed to patch resource")
			return
		}
	}

	sharedInformer.AddEventHandler(handlerFuncs)
	sharedInformer.Run(stream.Context().Done())

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

func createK8sClient() (dynamic.Interface, error) {
	var kubeconfigPath string

	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)

	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

func main() {
	port := flag.Int("port", 50051, "The server port")
	metadataStoreAddr := flag.String("mlmd-url", "", "The MLMD gRPC URL")
	flag.Parse()

	logger, err := logging.NewLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	k8sClient, err := createK8sClient()
	if err != nil {
		logger.Error(err, "failed to create k8s client")
		os.Exit(1)
	}


	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		logger.Error(err, "failed to listen")
		os.Exit(1)
	}


	conn, err := grpc.Dial(*metadataStoreAddr, grpc.WithInsecure())
	if err != nil {
		logger.Error(err, "failed to connect connect")
	}

	metadataStoreClient := ml_metadata.NewMetadataStoreServiceClient(conn)

	s := grpc.NewServer()
	generic.RegisterEventingServer(s, &eventingServer{
		k8sClient: k8sClient,
		logger: logger,
		metadataStore: &GrpcMetadataStore{
			MetadataStoreServiceClient: metadataStoreClient,
		},
	})
	logger.Info(fmt.Sprintf("server listening at %s", lis.Addr()))
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "failed to serve")
		os.Exit(1)
	}
}
