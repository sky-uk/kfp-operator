//go:build unit
// +build unit

package run_completion

import (
	"context"
	"errors"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func setWorkflowPhase(workflow *unstructured.Unstructured, phase argo.WorkflowPhase) {
	workflow.SetLabels(map[string]string{
		workflowPhaseLabel: string(phase),
	})
}

func setWorkflowEntryPoint(workflow *unstructured.Unstructured, entrypoint string) {
	workflow.Object = map[string]interface{}{
		"spec": map[string]interface{}{
			"entrypoint": entrypoint,
		},
	}
}

func setPipelineNameInSpec(workflow *unstructured.Unstructured, pipelineName string) {
	workflow.SetAnnotations(map[string]string{
		pipelineSpecAnnotationName: fmt.Sprintf(`{"name": "%s"}`, pipelineName),
	})
}

var _ = Context("Eventing Server", func() {
	Describe("jsonPatchPath", func() {
		It("concatenates path segments", func() {
			segment1 := randomString()
			segment2 := randomString()
			segment3 := randomString()

			expectedPath := fmt.Sprintf("/%s/%s/%s", segment1, segment2, segment3)

			Expect(jsonPatchPath(segment1, segment2, segment3)).To(Equal(expectedPath))
		})

		It("escapes '/'", func() {
			segment1 := randomString()
			segment2 := randomString()

			toBeEscaped := fmt.Sprintf("%s/%s", segment1, segment2)
			escaped := fmt.Sprintf("/%s~1%s", segment1, segment2)
			Expect(jsonPatchPath(toBeEscaped)).To(Equal(escaped))
		})

		It("escapes '~'", func() {
			segment1 := randomString()
			segment2 := randomString()

			toBeEscaped := fmt.Sprintf("%s~%s", segment1, segment2)
			escaped := fmt.Sprintf("/%s~0%s", segment1, segment2)
			Expect(jsonPatchPath(toBeEscaped)).To(Equal(escaped))
		})
	})

	Describe("runCompletionStatus", func() {
		It("returns false when the workflow has no status", func() {
			workflow := &unstructured.Unstructured{}
			_, hasFinished := runCompletionStatus(workflow)
			Expect(hasFinished).To(BeFalse())
		})
	})

	DescribeTable("runCompletionStatus",
		func(phase argo.WorkflowPhase) {
			workflow := &unstructured.Unstructured{}
			setWorkflowPhase(workflow, phase)
			_, hasFinished := runCompletionStatus(workflow)
			Expect(hasFinished).To(Equal(false))
		},
		Entry("unknown", argo.WorkflowUnknown),
		Entry("pending", argo.WorkflowPending),
		Entry("running", argo.WorkflowRunning),
	)

	DescribeTable("runCompletionStatus",
		func(phase argo.WorkflowPhase, expectedStatus RunCompletionStatus) {
			workflow := &unstructured.Unstructured{}
			setWorkflowPhase(workflow, phase)
			status, hasFinished := runCompletionStatus(workflow)
			Expect(status).To(Equal(expectedStatus))
			Expect(hasFinished).To(Equal(true))
		},
		Entry("error", argo.WorkflowError, Failed),
		Entry("failed", argo.WorkflowFailed, Failed),
		Entry("succeeded", argo.WorkflowSucceeded, Succeeded),
	)

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
			pipelineName := randomString()
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
			pipelineName := randomString()
			workflow := &unstructured.Unstructured{}
			setWorkflowEntryPoint(workflow, pipelineName)

			extractedName := getPipelineNameFromEntrypoint(workflow)
			Expect(extractedName).To(Equal(pipelineName))
		})
	})

	pipelineName := randomString()
	entrypoint := randomString()

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

			eventingServer := EventingServer{
				Logger: logr.Discard(),
			}

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("Doesn't emit an event when the workflow has no pipeline name", func() {
			workflow := &unstructured.Unstructured{}
			setWorkflowPhase(workflow, argo.WorkflowSucceeded)

			eventingServer := EventingServer{
				Logger: logr.Discard(),
			}

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors when the artifact store errors", func() {
			workflow := &unstructured.Unstructured{}
			setWorkflowPhase(workflow, argo.WorkflowSucceeded)
			setPipelineNameInSpec(workflow, randomString())

			mockMetadataStore := MockMetadataStore{}

			eventingServer := EventingServer{
				Logger:        logr.Discard(),
				MetadataStore: &mockMetadataStore,
			}

			expectedError := errors.New("an error occurred")
			mockMetadataStore.error(expectedError)

			event, err := eventingServer.eventForWorkflow(context.Background(), workflow)
			Expect(event).To(BeNil())
			Expect(err).To(Equal(expectedError))
		})
	})

	DescribeTable("eventForWorkflow", func(phase argo.WorkflowPhase) {
		workflow := &unstructured.Unstructured{}
		setWorkflowPhase(workflow, phase)
		setPipelineNameInSpec(workflow, randomString())
		workflow.SetName(randomString())

		mockMetadataStore := MockMetadataStore{}

		eventingServer := EventingServer{
			Logger:        logr.Discard(),
			MetadataStore: &mockMetadataStore,
		}

		artifacts := mockMetadataStore.returnArtifactForPipeline()
		event, err := eventingServer.eventForWorkflow(context.Background(), workflow)

		Expect(event.ServingModelArtifacts).To(Equal(artifacts))
		Expect(err).NotTo(HaveOccurred())
	},
		Entry("workflow succeeded", argo.WorkflowSucceeded),
		Entry("workflow failed", argo.WorkflowFailed),
		Entry("workflow errored", argo.WorkflowError),
	)
})
