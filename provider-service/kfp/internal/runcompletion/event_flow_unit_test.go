//go:build unit

package runcompletion

import (
	"context"
	"errors"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	common "github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Context("Eventing Flow", func() {
	Describe("getPipelineNameFromAnnotation", func() {
		It("returns empty when the workflow has no pipeline spec annotation", func() {
			workflow := &unstructured.Unstructured{}

			extractedName := getPipelineNameFromAnnotation(workflow)
			Expect(extractedName).To(BeEmpty())
		})

		It("returns empty when the workflow's pipeline spec annotation is invalid", func() {
			workflow := &unstructured.Unstructured{}
			workflow.SetAnnotations(map[string]string{
				pipelineSpecAnnotationName: fmt.Sprintf(`{invalid`),
			})

			extractedName := getPipelineNameFromAnnotation(workflow)
			Expect(extractedName).To(BeEmpty())
		})

		It("returns empty when the name is missing from workflow's spec annotation", func() {
			workflow := &unstructured.Unstructured{}
			workflow.SetAnnotations(map[string]string{
				pipelineSpecAnnotationName: fmt.Sprintf(`{}`),
			})

			extractedName := getPipelineNameFromAnnotation(workflow)
			Expect(extractedName).To(BeEmpty())
		})

		It("returns the pipeline's name when the workflow has a pipeline spec annotation with the pipeline name", func() {
			pipelineName := common.RandomString()
			workflow := &unstructured.Unstructured{}
			setPipelineNameInSpec(workflow, pipelineName)

			extractedName := getPipelineNameFromAnnotation(workflow)
			Expect(extractedName).To(Equal(pipelineName))
		})
	})

	Describe("eventForWorkflow", func() {
		It("Doesn't emit an event when the workflow has not finished", func() {
			workflow := &unstructured.Unstructured{}

			eventingServer := EventFlow{
				Logger: logr.Discard(),
			}

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("Doesn't emit an event when no ResourceReferences can be found and the workflow has no pipeline name", func() {
			workflow := &unstructured.Unstructured{}
			setWorkflowPhase(workflow, argo.WorkflowSucceeded)

			mockMetadataStore := mocks.MockMetadataStore{}
			mockKfpApi := mocks.MockKfpApi{}

			eventingServer := EventFlow{
				Logger:        logr.Discard(),
				MetadataStore: &mockMetadataStore,
				KfpApi:        &mockKfpApi,
			}

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors when the artifact store errors", func() {
			workflow := &unstructured.Unstructured{}
			workflow.SetName(common.RandomString())
			runId := common.RandomString()
			setWorkflowRunId(workflow, runId)
			setWorkflowPhase(workflow, argo.WorkflowSucceeded)
			setPipelineNameInSpec(workflow, common.RandomString())

			mockMetadataStore := mocks.MockMetadataStore{}
			mockKfpApi := mocks.MockKfpApi{}

			eventingServer := EventFlow{
				Logger:        logr.Discard(),
				MetadataStore: &mockMetadataStore,
				KfpApi:        &mockKfpApi,
			}

			mockKfpApi.On("GetResourceReferences", runId).Return(resource.RandomReferences(), nil)
			expectedError := errors.New("an error occurred")
			mockMetadataStore.On("GetArtifactsForRun", runId).Return(nil, expectedError)

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).To(Equal(expectedError))
		})

		It("errors when the KFP API errors", func() {
			workflow := &unstructured.Unstructured{}
			workflow.SetName(common.RandomString())
			runId := common.RandomString()
			setWorkflowRunId(workflow, runId)
			setWorkflowPhase(workflow, argo.WorkflowSucceeded)
			setPipelineNameInSpec(workflow, common.RandomString())

			mockMetadataStore := mocks.MockMetadataStore{}
			mockKfpApi := mocks.MockKfpApi{}

			eventingServer := EventFlow{
				Logger:        logr.Discard(),
				MetadataStore: &mockMetadataStore,
				KfpApi:        &mockKfpApi,
			}

			expectedError := errors.New("an error occurred")
			mockKfpApi.On("GetResourceReferences", runId).Return(nil, expectedError)

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).To(Equal(expectedError))
		})
	})

	DescribeTable("eventForWorkflow", func(phase argo.WorkflowPhase) {
		workflow := &unstructured.Unstructured{}
		runId := common.RandomString()
		setWorkflowRunId(workflow, runId)
		setWorkflowPhase(workflow, phase)
		setPipelineNameInSpec(workflow, common.RandomString())
		workflow.SetName(common.RandomString())

		mockMetadataStore := mocks.MockMetadataStore{}
		mockKfpApi := mocks.MockKfpApi{}

		expectedComponents := []common.PipelineComponent{
			{
				Name: "test-component",
				ComponentArtifacts: []common.ComponentArtifact{
					{
						Name: "test-artifact",
						Artifacts: []common.ComponentArtifactInstance{
							{
								Uri:      "gs://test/artifact",
								Metadata: map[string]any{"key": "value"},
							},
						},
					},
				},
			},
		}
		mockMetadataStore.On("GetArtifactsForRun", runId).Return(expectedComponents, nil)
		resourceReferences := resource.RandomReferences()
		mockKfpApi.On("GetResourceReferences", runId).Return(resourceReferences, nil)

		eventingServer := EventFlow{
			Logger:        logr.Discard(),
			MetadataStore: &mockMetadataStore,
			KfpApi:        &mockKfpApi,
			ProviderConfig: config.Config{
				ProviderName: common.NamespacedName{Namespace: "default", Name: "kfp"},
			},
		}

		event, err := eventingServer.eventForWorkflow(context.Background(), workflow)

		Expect(*event.RunConfigurationName).To(Equal(resourceReferences.RunConfigurationName))
		Expect(*event.RunName).To(Equal(resourceReferences.RunName))
		Expect(event.Provider.Name).To(Equal(eventingServer.ProviderConfig.ProviderName.Name))
		Expect(event.Provider.Namespace).To(Equal(eventingServer.ProviderConfig.ProviderName.Namespace))
		Expect(event.PipelineComponents).To(Equal(expectedComponents))
		Expect(err).NotTo(HaveOccurred())
	},
		Entry("workflow succeeded", argo.WorkflowSucceeded),
		Entry("workflow failed", argo.WorkflowFailed),
		Entry("workflow errored", argo.WorkflowError),
	)

})

func setPipelineNameInSpec(workflow *unstructured.Unstructured, pipelineName string) {
	workflow.SetAnnotations(map[string]string{
		pipelineSpecAnnotationName: fmt.Sprintf(`{"name": "%s"}`, pipelineName),
	})
}

func setWorkflowEntryPoint(workflow *unstructured.Unstructured, entrypoint string) {
	workflow.Object = map[string]any{
		"spec": map[string]any{
			"entrypoint": entrypoint,
		},
	}
}

func setWorkflowRunId(
	workflow *unstructured.Unstructured,
	runId string,
) {
	labels := workflow.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[pipelineRunIdLabel] = runId
	workflow.SetLabels(labels)
}

func setWorkflowPhase(
	workflow *unstructured.Unstructured,
	phase argo.WorkflowPhase,
) {
	labels := workflow.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[workflowPhaseLabel] = string(phase)
	workflow.SetLabels(labels)
}
