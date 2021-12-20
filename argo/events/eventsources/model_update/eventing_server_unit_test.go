//go:build unit
// +build unit

package model_update

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Context("Eventing Server", func() {
	Describe("jsonPatchPath", func() {
		It("concatenates path segments", func() {
			segment1 := randomString()
			segment2 := randomString()
			segment3 := randomString()

			joinedPath := jsonPatchPath(segment1, segment2, segment3)

			Expect(fmt.Sprintf("/%s/%s/%s", segment1, segment2, segment3)).To(Equal(joinedPath))
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

	Describe("workflowHasSucceeded", func() {
		When("The workflow has no status", func() {
			It("Returns false", func() {
				workflow := &unstructured.Unstructured{}
				Expect(workflowHasSucceeded(workflow)).To(BeFalse())
			})
		})

		When("The workflow has not succeeded", func() {
			DescribeTable("Returns false",
				func(phase argo.WorkflowPhase) {
					workflow := &unstructured.Unstructured{}
					workflow.SetLabels(map[string]string{
						workflowPhaseLabel: string(phase),
					})
					Expect(workflowHasSucceeded(workflow)).To(BeFalse())
				},
				Entry("unknown", argo.WorkflowUnknown),
				Entry("pending", argo.WorkflowPending),
				Entry("running", argo.WorkflowRunning),
				Entry("error", argo.WorkflowError),
				Entry("failed", argo.WorkflowFailed),
			)
		})

		When("The workflow has succeeded", func() {
			It("Returns true", func() {
				workflow := &unstructured.Unstructured{}
				workflow.SetLabels(map[string]string{
					workflowPhaseLabel: string(argo.WorkflowSucceeded),
				})
				Expect(workflowHasSucceeded(workflow)).To(BeTrue())
			})
		})
	})

	Describe("getPipelineName", func() {
		When("The workflow has no pipeline spec annotation", func() {
			It("Returns false", func() {
				workflow := &unstructured.Unstructured{}

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow's pipeline spec annotation is invalid", func() {
			It("Errors", func() {
				workflow := &unstructured.Unstructured{}
				workflow.SetAnnotations(map[string]string{
					pipelineSpecAnnotationName: fmt.Sprintf(`{invalid`),
				})

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow's spec annotation is empty", func() {
			It("Errors", func() {
				workflow := &unstructured.Unstructured{}
				workflow.SetAnnotations(map[string]string{
					pipelineSpecAnnotationName: fmt.Sprintf(`{"name": ""}`),
				})

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow's spec annotation is missing", func() {
			It("Errors", func() {
				workflow := &unstructured.Unstructured{}
				workflow.SetAnnotations(map[string]string{
					pipelineSpecAnnotationName: fmt.Sprintf(`{}`),
				})

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow has a pipeline spec annotation with the pipeline name", func() {
			It("returns the name", func() {
				pipelineName := randomString()
				workflow := &unstructured.Unstructured{}
				workflow.SetAnnotations(map[string]string{
					pipelineSpecAnnotationName: fmt.Sprintf(`{"name": "%s"}`, pipelineName),
				})

				extractedName, err := getPipelineName(workflow)
				Expect(err).NotTo(HaveOccurred())
				Expect(extractedName).To(Equal(pipelineName))
			})
		})
	})
})
