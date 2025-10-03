//go:build unit

package runcompletion

import (
	"context"
	"errors"
	"fmt"
	common "github.com/sky-uk/kfp-operator/pkg/common"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	Describe("getPipelineNameFromEntrypoint", func() {
		It("returns empty when the workflow has no entrypoint", func() {
			workflow := &unstructured.Unstructured{}

			extractedName := getPipelineNameFromEntrypoint(workflow)
			Expect(extractedName).To(BeEmpty())
		})

		It("returns empty when the workflow's entrypoint is empty'", func() {
			workflow := &unstructured.Unstructured{}
			setWorkflowEntryPoint(workflow, "")

			extractedName := getPipelineNameFromEntrypoint(workflow)
			Expect(extractedName).To(BeEmpty())
		})

		It("returns the pipeline's name when the workflow has an entrypoint'", func() {
			pipelineName := common.RandomString()
			workflow := &unstructured.Unstructured{}
			setWorkflowEntryPoint(workflow, pipelineName)

			extractedName := getPipelineNameFromEntrypoint(workflow)
			Expect(extractedName).To(Equal(pipelineName))
		})
	})

	pipelineName := common.RandomString()
	entrypoint := common.RandomString()

	DescribeTable("getPipelineName", func(annotationValue string, entrypoint string, expected string) {
		workflow := &unstructured.Unstructured{}
		setWorkflowEntryPoint(workflow, entrypoint)
		setPipelineNameInSpec(workflow, annotationValue)

		extractedName := getPipelineName(workflow)
		Expect(extractedName).To(Equal(expected))
	},
		Entry("Returns empty when none is present", "", "", ""),
		Entry("Returns the annotation value", pipelineName, "", pipelineName),
		Entry("Falls back to the entrypoint", "", entrypoint, entrypoint),
		Entry("Prefers the annotation value", pipelineName, entrypoint, pipelineName),
	)

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
			mockMetadataStore.Error(expectedError)

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).To(Equal(expectedError))
		})

		It("errors when the KFP API errors", func() {
			workflow := &unstructured.Unstructured{}
			workflow.SetName(common.RandomString())
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
			mockKfpApi.Error(expectedError)

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).To(Equal(expectedError))
		})
	})

	DescribeTable("eventForWorkflow", func(phase argo.WorkflowPhase) {
		workflow := &unstructured.Unstructured{}
		setWorkflowPhase(workflow, phase)
		setPipelineNameInSpec(workflow, common.RandomString())
		workflow.SetName(common.RandomString())

		mockMetadataStore := mocks.MockMetadataStore{}
		mockKfpApi := mocks.MockKfpApi{}

		eventingServer := EventFlow{
			Logger:        logr.Discard(),
			MetadataStore: &mockMetadataStore,
			KfpApi:        &mockKfpApi,
			ProviderConfig: config.Config{
				Name: "kfp",
			},
		}

		artifacts := mockMetadataStore.ReturnArtifactForPipeline()
		resourceReferences := mockKfpApi.ReturnResourceReferencesForRun()
		event, err := eventingServer.eventForWorkflow(context.Background(), workflow)

		Expect(event.ServingModelArtifacts).To(Equal(artifacts))
		Expect(*event.RunConfigurationName).To(Equal(resourceReferences.RunConfigurationName))
		Expect(*event.RunName).To(Equal(resourceReferences.RunName))
		Expect(event.Provider).To(Equal(eventingServer.ProviderConfig.Name))
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
	workflow.Object = map[string]interface{}{
		"spec": map[string]interface{}{
			"entrypoint": entrypoint,
		},
	}
}

func setWorkflowPhase(workflow *unstructured.Unstructured, phase argo.WorkflowPhase) {
	workflow.SetLabels(map[string]string{
		workflowPhaseLabel: string(phase),
	})
}
