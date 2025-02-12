//go:build decoupled

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"path/filepath"
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sinks"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/streams/sources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	mockMetadataStore mocks.MockMetadataStore
	mockKfpApi        mocks.MockKfpApi
	k8sClient         dynamic.Interface
	eventSource       *sources.WorkflowSource
	webhookSink       *sinks.WebhookSink
	eventFlow         streams.Flow[pkg.StreamMessage[*unstructured.Unstructured], pkg.StreamMessage[*common.RunCompletionEventData], error]
	eventData         common.RunCompletionEventData
	numberOfEvents    int
)

var (
	argoWorkflowsGvr = schema.GroupVersionResource{
		Group:    workflow.Group,
		Version:  workflow.Version,
		Resource: workflow.WorkflowPlural,
	}
)

const (
	defaultNamespace = "default"
	providerName     = "kfp"
	webhookUrl       = "/operator-webhook"
)

var _ = BeforeSuite(func() {
	ctx := context.Background()

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

	mockMetadataStore = mocks.MockMetadataStore{}
	mockKfpApi = mocks.MockKfpApi{}

	eventSource, err = sources.NewWorkflowSource(ctx, defaultNamespace, pkg.K8sClient{Client: k8sClient})
	Expect(err).ToNot(HaveOccurred())

	config := &config.KfpProviderConfig{
		Name: providerName,
	}

	eventSource, err = sources.NewWorkflowSource(context.Background(), defaultNamespace, pkg.K8sClient{Client: k8sClient})
	Expect(err).ToNot(HaveOccurred())

	eventFlow, err = NewEventFlow(context.Background(), *config, &mockKfpApi, &mockMetadataStore)
	Expect(err).ToNot(HaveOccurred())
})

var _ = BeforeEach(func() {
	numberOfEvents = 0
})

func WithTestContext(fun func(context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)
	defer cancel()

	Expect(deleteAllWorkflows(ctx)).To(Succeed())
	mockMetadataStore.Reset()
	mockKfpApi.Reset()
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
	webhookSink = sinks.NewWebhookSink(ctx, client, webhookUrl, make(chan pkg.StreamMessage[*common.RunCompletionEventData]))

	go func() {
		eventFlow.From(eventSource).To(webhookSink)
	}()

	fun(ctx)
}

func getEventData() common.RunCompletionEventData { return eventData }

var _ = Describe("Run completion eventsource", Serial, func() {
	When("A pipeline run succeeds and a model has been pushed", func() {
		It("Triggers an event with serving model artifacts", func() {
			WithTestContext(func(ctx context.Context) {
				resourceReferences := mockKfpApi.ReturnResourceReferencesForRun()
				pipelineName, err := resourceReferences.PipelineName.String()
				Expect(err).ToNot(HaveOccurred())

				runId := common.RandomString()

				servingModelArtifacts := mockMetadataStore.ReturnArtifactForPipeline()

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
				resourceReferences := mockKfpApi.ReturnResourceReferencesForRun()
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
				resourceReferences := mockKfpApi.ReturnResourceReferencesForRun()
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
	wf := unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{},
		},
	}
	wf.SetGroupVersionKind(argo.WorkflowSchemaGroupVersionKind)
	wf.SetName(rand.String(10))
	wf.SetLabels(map[string]string{
		workflowPhaseLabel: string(phase),
		pipelineRunIdLabel: runId,
	})
	wf.SetAnnotations(map[string]string{
		pipelineSpecAnnotationName: fmt.Sprintf(`{"name": "%s"}`, pipelineName),
	})

	return k8sClient.Resource(argoWorkflowsGvr).Namespace(defaultNamespace).Create(ctx, &wf, metav1.CreateOptions{})
}

func createAndTriggerPhaseUpdate(ctx context.Context, pipelineName string, runId string, from argo.WorkflowPhase, to argo.WorkflowPhase) (*unstructured.Unstructured, error) {
	wf, err := createWorkflowInPhase(ctx, pipelineName, runId, from)
	if err != nil {
		return nil, err
	}

	return updatePhase(ctx, wf.GetName(), to)
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
