//go:build decoupled

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/publisher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	mockMetadataStore MockMetadataStore
	mockKfpApi        MockKfpApi
	k8sClient         dynamic.Interface
	eventSource       *KfpEventSource
	webhookSink       *publisher.HttpWebhookSink
	eventData         common.RunCompletionEventData
	numberOfEvents    int
)

const (
	defaultNamespace = "default"
	providerName     = "kfp"
	webhookUrl       = "/operator-webhook"
)

func TestModelUpdateEventSourceDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Completion EventSource Decoupled Suite")
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

func expectedNumberOfEventsOccurred(ctx context.Context, expectedNumberOfEvents int) {
	const Marker = "marker"

	_, err := createAndTriggerPhaseUpdate(ctx, "p-name", Marker, argo.WorkflowRunning, argo.WorkflowSucceeded)

	Expect(err).ToNot(HaveOccurred())
	Eventually(func() string { return getEventData().RunId }).Should(Equal(Marker))
	Eventually(func() int { return numberOfEvents }).Should(Equal(expectedNumberOfEvents + 1))
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

	mockMetadataStore = MockMetadataStore{}
	mockKfpApi = MockKfpApi{}

	config := &KfpProviderConfig{
		Name: providerName,
	}

	eventFlow := KfpEventFlow{
		ProviderConfig: *config,
		MetadataStore:  &mockMetadataStore,
		KfpApi:         &mockKfpApi,
		Logger:         logr.Discard(),
		context:        context.Background(),
		in:             make(chan pkg.StreamMessage[*unstructured.Unstructured]),
		out:            make(chan pkg.StreamMessage[*common.RunCompletionEventData]),
		errorOut:       make(chan error),
	}

	eventSource = &KfpEventSource{
		K8sClient:                        pkg.K8sClient{Client: k8sClient},
		RunCompletionEventConversionFlow: &eventFlow,
		Logger:                           logr.Discard(),
		out:                              make(chan pkg.StreamMessage[*unstructured.Unstructured]),
	}

	ctx := context.Background()

	go startEventSource(ctx)
})

func startEventSource(ctx context.Context) {
	err := eventSource.start(ctx, defaultNamespace)
	if err != nil {
		return
	}
}

var _ = AfterEach(func() {
	numberOfEvents = 0
})

func WithTestContext(fun func(context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)
	defer cancel()

	Expect(deleteAllWorkflows(ctx)).To(Succeed())
	mockMetadataStore.reset()
	mockKfpApi.reset()
	eventSource.out = make(chan pkg.StreamMessage[*unstructured.Unstructured])
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	httpmock.RegisterResponder(
		"POST",
		webhookUrl,
		func(req *http.Request) (*http.Response, error) {
			numberOfEvents++
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return httpmock.NewStringResponse(503, "failed to read body"), err
			}
			err = json.Unmarshal(body, &eventData)
			if err != nil {
				return httpmock.NewStringResponse(503, "failed to unmarshal body"), err
			}
			return httpmock.NewStringResponse(200, ""), nil
		},
	)
	webhookSink = publisher.NewHttpWebhookSink(ctx, webhookUrl, client, make(chan pkg.StreamMessage[*common.RunCompletionEventData]))

	go eventSource.RunCompletionEventConversionFlow.From(eventSource).To(webhookSink)
	fun(ctx)
}

func getEventData() common.RunCompletionEventData { return eventData }

var _ = Describe("Run completion eventsource", Serial, func() {
	When("A pipeline run succeeds and a model has been pushed", func() {
		It("Triggers an event with serving model artifacts", func() {
			WithTestContext(func(ctx context.Context) {
				resourceReferences := mockKfpApi.returnResourceReferencesForRun()
				pipelineName, err := resourceReferences.PipelineName.String()
				Expect(err).ToNot(HaveOccurred())

				runId := common.RandomString()

				servingModelArtifacts := mockMetadataStore.returnArtifactForPipeline()

				workflow, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				expectedRced := common.RunCompletionEventData{
					Status:                common.RunCompletionStatuses.Succeeded,
					PipelineName:          resourceReferences.PipelineName,
					RunConfigurationName:  resourceReferences.RunConfigurationName.NonEmptyPtr(),
					RunName:               resourceReferences.RunName.NonEmptyPtr(),
					RunId:                 runId,
					ServingModelArtifacts: servingModelArtifacts,
					Provider:              providerName,
				}

				Eventually(getEventData).Should(Equal(expectedRced))
				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				Expect(triggerUpdate(ctx, workflow.GetName())).To(Succeed())
				expectedNumberOfEventsOccurred(ctx, 1)
			})
		})
	})

	When("A pipeline run succeeds and no model has been pushed and no RunConfiguration is found", func() {
		It("Triggers an event without a serving model artifacts", func() {
			WithTestContext(func(ctx context.Context) {
				resourceReferences := mockKfpApi.returnResourceReferencesForRun()
				pipelineName, err := resourceReferences.PipelineName.String()
				Expect(err).ToNot(HaveOccurred())

				runId := common.RandomString()

				workflow, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
				Expect(err).NotTo(HaveOccurred())

				expectedRced := common.RunCompletionEventData{
					Status:               common.RunCompletionStatuses.Succeeded,
					PipelineName:         resourceReferences.PipelineName,
					RunId:                runId,
					RunConfigurationName: resourceReferences.RunConfigurationName.NonEmptyPtr(),
					RunName:              resourceReferences.RunName.NonEmptyPtr(),
					Provider:             providerName,
				}

				Eventually(getEventData).Should(Equal(expectedRced))
				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				expectedNumberOfEventsOccurred(ctx, 1)
			})
		})
	})

	When("A pipeline run fails", func() {
		It("Triggers an event", func() {
			WithTestContext(func(ctx context.Context) {
				resourceReferences := mockKfpApi.returnResourceReferencesForRun()
				pipelineName, err := resourceReferences.PipelineName.String()
				Expect(err).ToNot(HaveOccurred())

				runId := common.RandomString()

				workflow, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowFailed)
				Expect(err).NotTo(HaveOccurred())

				expectedRced := common.RunCompletionEventData{
					Status:               common.RunCompletionStatuses.Failed,
					PipelineName:         resourceReferences.PipelineName,
					RunId:                runId,
					RunConfigurationName: resourceReferences.RunConfigurationName.NonEmptyPtr(),
					RunName:              resourceReferences.RunName.NonEmptyPtr(),
					Provider:             providerName,
				}

				Eventually(getEventData).Should(Equal(expectedRced))
				Eventually(func(g Gomega) {
					g.Expect(workflowLabel(ctx, workflow.GetName(), workflowUpdateTriggeredLabel)).To(Equal("true"))
				}).Should(Succeed())

				Expect(triggerUpdate(ctx, workflow.GetName())).To(Succeed())
				expectedNumberOfEventsOccurred(ctx, 1)
			})
		})
	})

	//When("A pipeline run finishes before the stream is started", func() {
	//	It("Catches up and triggers an event", func() {
	//		WithTestContext(func(ctx context.Context) {
	//			pipelineName := common.RandomString()
	//			runId := common.RandomString()
	//
	//			_, err := createAndTriggerPhaseUpdate(ctx, pipelineName, runId, argo.WorkflowRunning, argo.WorkflowSucceeded)
	//			Expect(err).NotTo(HaveOccurred())
	//
	//			expectedRced := common.RunCompletionEventData{
	//				Status:       common.RunCompletionStatuses.Succeeded,
	//				PipelineName: common.NamespacedName{Name: pipelineName},
	//				RunId:        runId,
	//				Provider:     providerName,
	//			}
	//			Eventually(getEventData).Should(Equal(expectedRced))
	//		})
	//	})
	//})

	When("A pipeline run doesn't finish", func() {
		It("Does not trigger an event", func() {
			WithTestContext(func(ctx context.Context) {
				_, err := createAndTriggerPhaseUpdate(ctx, common.RandomString(), common.RandomString(), argo.WorkflowPending, argo.WorkflowRunning)
				Expect(err).NotTo(HaveOccurred())

				expectedNumberOfEventsOccurred(ctx, 0)
			})
		})
	})
})
