//go:build decoupled

package kfp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-events/eventsources/sources/generic"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/dynamic"
	"net"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
	"time"
)

var (
	mockMetadataStore MockMetadataStore
	mockKfpApi        MockKfpApi
	server            *grpc.Server
	k8sClient         dynamic.Interface
	clientConn        *grpc.ClientConn
)

const (
	defaultNamespace = "default"
)

func TestModelUpdateEventSourceDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Completion EventSource Decoupled Suite")
}

func startClient(ctx context.Context) (generic.Eventing_StartEventSourceClient, error) {
	eventSourceConfig, err := yaml.Marshal(&KfpEventSourceConfig{
		KfpNamespace: defaultNamespace,
	})
	if err != nil {
		return nil, err
	}

	eventingClient := generic.NewEventingClient(clientConn)

	return eventingClient.StartEventSource(ctx, &generic.EventSource{
		Name:   rand.String(10),
		Config: eventSourceConfig,
	})
}

func deleteAllWorkflows(ctx context.Context) error {
	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

func workflowLabel(ctx context.Context, name string, key string) (string, error) {
	resource, err := k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return resource.GetLabels()[key], nil
}

func updateLabel(ctx context.Context, name string, key string, value string) (*unstructured.Unstructured, error) {
	resource, err := k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	updatedLabels := resource.GetLabels()
	updatedLabels[key] = value
	resource.SetLabels(updatedLabels)

	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Update(ctx, resource, metav1.UpdateOptions{})
}

func updatePhase(ctx context.Context, name string, phase argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	return updateLabel(ctx, name, workflowPhaseLabel, string(phase))
}

func createWorkflowInPhase(ctx context.Context, pipelineName string, runId string, phase argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	workflow := unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{},
		},
	}
	workflow.SetGroupVersionKind(argo.WorkflowSchemaGroupVersionKind)
	workflow.SetName(rand.String(10))
	workflow.SetLabels(map[string]string{
		workflowPhaseLabel: string(phase),
		pipelineRunIdLabel: runId,
	})
	workflow.SetAnnotations(map[string]string{
		pipelineSpecAnnotationName: fmt.Sprintf(`{"name": "%s"}`, pipelineName),
	})

	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Create(ctx, &workflow, metav1.CreateOptions{})
}

func createAndTriggerPhaseUpdate(ctx context.Context, pipelineName string, runId string, from argo.WorkflowPhase, to argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	workflow, err := createWorkflowInPhase(ctx, pipelineName, runId, from)
	if err != nil {
		return nil, err
	}

	return updatePhase(ctx, workflow.GetName(), to)
}

func triggerUpdate(ctx context.Context, name string) error {
	_, err := updateLabel(ctx, name, rand.String(5), rand.String(5))

	return err
}

func furtherEvents(ctx context.Context, stream generic.Eventing_StartEventSourceClient) error {
	const Marker = "marker"

	_, err := createAndTriggerPhaseUpdate(ctx, common.RandomString(), Marker, argo.WorkflowRunning, argo.WorkflowSucceeded)
	if err != nil {
		return err
	}

	event, err := stream.Recv()
	if err != nil {
		return err
	}

	actualEvent := common.RunCompletionEvent{}
	err = json.Unmarshal(event.Payload, &actualEvent)
	if err != nil {
		return err
	}

	if actualEvent.RunId != Marker {
		return fmt.Errorf("unexpected event: %+v", event)
	}

	return nil
}

var _ = BeforeSuite(func() {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "", "config", "crd", "external"),
		},
		ErrorIfCRDPathMissing: true,
	}

	k8sConfig, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	k8sClient, err = dynamic.NewForConfig(k8sConfig)
	Expect(err).NotTo(HaveOccurred())

	Expect(err).NotTo(HaveOccurred())

	lis, err := net.Listen("tcp", "127.0.0.1:")
	Expect(err).NotTo(HaveOccurred())

	mockMetadataStore = MockMetadataStore{}
	mockKfpApi = MockKfpApi{}
	server = grpc.NewServer()

	generic.RegisterEventingServer(server, &KfpEventingServer{
		K8sClient:     k8sClient,
		Logger:        logr.Discard(),
		MetadataStore: &mockMetadataStore,
		KfpApi:        &mockKfpApi,
	})

	go server.Serve(lis)

	clientConn, err = grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	server.Stop()
	clientConn.Close()
})

func WithTestContext(fun func(context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer cancel()

	Expect(deleteAllWorkflows(ctx)).To(Succeed())
	mockMetadataStore.reset()
	mockKfpApi.reset()

	fun(ctx)
}

var _ = Describe("Run completion eventsource", Serial, func() {
	When("A pipeline run succeeds and a model has been pushed", func() {
		It("Triggers an event with serving model artifacts", func() {
			WithTestContext(func(ctx context.Context) {
				stream, err := startClient(ctx)
				pipelineName := common.RandomString()
				runId := common.RandomString()
				servingModelArtifacts := mockMetadataStore.returnArtifactForPipeline()
				resourceReferences := mockKfpApi.returnResourceReferencesForRun()

				Expect(err).NotTo(HaveOccurred())

				workflow, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(common.RunCompletionEventName))

				expectedEvent := common.RunCompletionEvent{
					Status:                common.RunCompletionStatuses.Succeeded,
					PipelineName:          resourceReferences.PipelineName,
					RunConfigurationName:  resourceReferences.RunConfigurationName.NonEmptyPtr(),
					RunName:               resourceReferences.RunName.NonEmptyPtr(),
					RunId:                 runId,
					ServingModelArtifacts: servingModelArtifacts,
				}
				actualEvent := common.RunCompletionEvent{}
				err = json.Unmarshal(event.Payload, &actualEvent)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualEvent).To(Equal(expectedEvent))

				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				Expect(triggerUpdate(ctx, workflow.GetName())).To(Succeed())
				Expect(furtherEvents(ctx, stream)).NotTo(HaveOccurred())
			})
		})
	})

	When("A pipeline run succeeds and no model has been pushed and no RunConfiguration is found", func() {
		It("Triggers an event without a serving model artifacts", func() {
			WithTestContext(func(ctx context.Context) {
				stream, err := startClient(ctx)
				pipelineName := common.RandomString()
				runId := common.RandomString()

				Expect(err).NotTo(HaveOccurred())

				workflow, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(common.RunCompletionEventName))

				expectedEvent := common.RunCompletionEvent{
					Status:       common.RunCompletionStatuses.Succeeded,
					PipelineName: common.NamespacedName{Name: pipelineName},
					RunId:        runId,
				}
				actualEvent := common.RunCompletionEvent{}
				err = json.Unmarshal(event.Payload, &actualEvent)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualEvent).To(Equal(expectedEvent))

				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				Expect(triggerUpdate(ctx, workflow.GetName())).To(Succeed())
				Expect(furtherEvents(ctx, stream)).NotTo(HaveOccurred())
			})
		})
	})

	When("A pipeline run fails", func() {
		It("Triggers an event", func() {
			WithTestContext(func(ctx context.Context) {
				stream, err := startClient(ctx)
				pipelineName := common.RandomString()
				runId := common.RandomString()

				Expect(err).NotTo(HaveOccurred())

				workflow, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowFailed)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(common.RunCompletionEventName))

				expectedEvent := common.RunCompletionEvent{
					Status:       common.RunCompletionStatuses.Failed,
					PipelineName: common.NamespacedName{Name: pipelineName},
					RunId:        runId,
				}
				actualEvent := common.RunCompletionEvent{}
				err = json.Unmarshal(event.Payload, &actualEvent)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualEvent).To(Equal(expectedEvent))

				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				Expect(triggerUpdate(ctx, workflow.GetName())).To(Succeed())
				Expect(furtherEvents(ctx, stream)).NotTo(HaveOccurred())
			})
		})
	})

	When("A pipeline run finishes before the stream is started", func() {
		It("Catches up and triggers an event", func() {
			WithTestContext(func(ctx context.Context) {
				pipelineName := common.RandomString()
				runId := common.RandomString()

				_, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(common.RunCompletionEventName))

				expectedEvent := common.RunCompletionEvent{
					Status:       common.RunCompletionStatuses.Succeeded,
					PipelineName: common.NamespacedName{Name: pipelineName},
					RunId:        runId,
				}
				actualEvent := common.RunCompletionEvent{}
				err = json.Unmarshal(event.Payload, &actualEvent)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualEvent).To(Equal(expectedEvent))
			})
		})
	})

	When("A pipeline run doesn't finish", func() {
		It("Does not trigger an event", func() {
			WithTestContext(func(ctx context.Context) {
				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				_, err = createAndTriggerPhaseUpdate(ctx, common.RandomString(), common.RandomString(), argo.WorkflowPending, argo.WorkflowRunning)
				Expect(err).NotTo(HaveOccurred())

				Expect(furtherEvents(ctx, stream)).NotTo(HaveOccurred())
			})
		})
	})

	When("A pipeline run succeeds but the artifact store is unavailable", func() {
		It("Retries", func() {
			WithTestContext(func(ctx context.Context) {
				pipelineName := common.RandomString()
				runId := common.RandomString()

				mockMetadataStore.error(errors.New("error calling metadata store"))

				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				_, err = createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				_, err = stream.Recv()
				Expect(err).To(HaveOccurred())

				servingModelArtifacts := mockMetadataStore.returnArtifactForPipeline()

				stream, err = startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(common.RunCompletionEventName))

				expectedEvent := common.RunCompletionEvent{
					Status:                common.RunCompletionStatuses.Succeeded,
					PipelineName:          common.NamespacedName{Name: pipelineName},
					RunId:                 runId,
					ServingModelArtifacts: servingModelArtifacts,
				}
				actualEvent := common.RunCompletionEvent{}
				err = json.Unmarshal(event.Payload, &actualEvent)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualEvent).To(Equal(expectedEvent))
			})
		})
	})

	When("A pipeline run succeeds but the KFP API is unavailable", func() {
		It("Retries", func() {
			WithTestContext(func(ctx context.Context) {
				pipelineName := common.RandomString()
				runId := common.RandomString()

				mockKfpApi.error(errors.New("error calling KFP API"))

				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				_, err = createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				_, err = stream.Recv()
				Expect(err).To(HaveOccurred())

				resourceReferences := mockKfpApi.returnResourceReferencesForRun()

				stream, err = startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(common.RunCompletionEventName))

				expectedEvent := common.RunCompletionEvent{
					Status:               common.RunCompletionStatuses.Succeeded,
					PipelineName:         resourceReferences.PipelineName,
					RunConfigurationName: resourceReferences.RunConfigurationName.NonEmptyPtr(),
					RunName:              resourceReferences.RunName.NonEmptyPtr(),
					RunId:                runId,
				}
				actualEvent := common.RunCompletionEvent{}
				err = json.Unmarshal(event.Payload, &actualEvent)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualEvent).To(Equal(expectedEvent))
			})
		})
	})
})
