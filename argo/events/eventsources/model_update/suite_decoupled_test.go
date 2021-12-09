//go:build decoupled
// +build decoupled

package main

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/dynamic"
	"net"
	"path/filepath"
	"pipelines.kubeflow.org/events/eventsources/generic"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
	"time"
)

var (
	server *grpc.Server
	k8sClient dynamic.Interface
	clientConn *grpc.ClientConn
)

const (
	defaultNamespace = "default"
)

func TestModelUpdateEventSourceDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Update EventSource Decoupled Suite")
}

func startClient(ctx context.Context) (generic.Eventing_StartEventSourceClient, error) {
	eventSourceConfig, err := yaml.Marshal(&EventSourceConfig{
		KfpNamespace: defaultNamespace,
	})
	if err != nil {
		return nil, err
	}

	eventingClient := generic.NewEventingClient(clientConn)

	return eventingClient.StartEventSource(ctx, &generic.EventSource{
		Name: rand.String(10),
		Config: eventSourceConfig,
	})
}

func deleteAllWorkflows(ctx context.Context) error {
	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).DeleteCollection(ctx, v1.DeleteOptions{}, v1.ListOptions{})
}

func workflowLabel(ctx context.Context, name string, key string) (string, error) {
	resource, err := k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return "", err
	}

	return resource.GetLabels()[key], nil
}

func updateLabel(ctx context.Context, name string, key string, value string) (*unstructured.Unstructured, error) {
	resource, err := k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	updatedLabels := resource.GetLabels()
	updatedLabels[key] = value
	resource.SetLabels(updatedLabels)

	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Update(ctx, resource, v1.UpdateOptions{})
}

func updatePhase(ctx context.Context, name string, phase argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	return updateLabel(ctx, name, workflowPhaseLabel, string(phase))
}

func createWorkflowInPhase(ctx context.Context, phase argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	workflow := unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{

			},
		},
	}
	workflow.SetGroupVersionKind(argo.WorkflowSchemaGroupVersionKind)
	workflow.SetName(rand.String(10))
	workflow.SetLabels(map[string]string{
		workflowPhaseLabel: string(phase),
		kfpSdkVersionLabel: rand.String(5),
	})

	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Create(ctx, &workflow, v1.CreateOptions{})
}

func createAndTriggerPhaseUpdate(ctx context.Context, from argo.WorkflowPhase, to argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	workflow, err := createWorkflowInPhase(ctx, from)
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
	workflow, err := createAndTriggerPhaseUpdate(ctx, argo.WorkflowRunning, argo.WorkflowSucceeded)
	if err != nil {
		return err
	}

	event, err := stream.Recv()
	if err != nil {
		return err
	}

	if string(event.Payload) != workflow.GetName() {
		return fmt.Errorf("unexpected event: %+v", event)
	}

	return nil
}

var _ = BeforeSuite(func() {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "..", "", "config", "crd", "external"),
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

	server = grpc.NewServer()
	generic.RegisterEventingServer(server, &eventingServer{
		k8sClient: k8sClient,
		logger: logr.Discard(),
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

	fun(ctx)
}

var _ = Describe("Model update eventsource", func() {
	When("A pipeline run succeeds", func() {
		It("Triggers an event", func() {
			WithTestContext(func(ctx context.Context) {
				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				workflow, err := createAndTriggerPhaseUpdate(ctx,  argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(modelUpdateEventName))
				Expect(event.Payload).To(Equal([]byte(workflow.GetName())))

				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				Expect(triggerUpdate(ctx, workflow.GetName())).To(Succeed())
				Expect(furtherEvents(ctx, stream)).NotTo(HaveOccurred())
			})
		})
	})

	When("A pipeline run succeeds before the stream is started", func() {
		It("Catches up and triggers an event", func() {
			WithTestContext(func(ctx context.Context) {
				workflow, err := createAndTriggerPhaseUpdate(ctx,  argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				event, err := stream.Recv()
				Expect(err).NotTo(HaveOccurred())

				Expect(event.Name).To(Equal(modelUpdateEventName))
				Expect(event.Payload).To(Equal([]byte(workflow.GetName())))
			})
		})
	})

	When("A pipeline run does not succeed", func() {
		It("Does not trigger an event", func() {
			WithTestContext(func(ctx context.Context) {
				stream, err := startClient(ctx)
				Expect(err).NotTo(HaveOccurred())

				_, err = createAndTriggerPhaseUpdate(ctx,  argo.WorkflowPending, argo.WorkflowRunning)
				Expect(err).NotTo(HaveOccurred())

				Expect(furtherEvents(ctx, stream)).NotTo(HaveOccurred())
			})
		})
	})
})
