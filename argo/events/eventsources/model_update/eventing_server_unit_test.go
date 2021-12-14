//go:build unit
// +build unit

package main

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Context("Eventing Server", func() {
	Describe("getPipelineName", func() {
		When("The workflow has no pipeline spec annotation", func() {
			It("Errors", func() {
				workflow := &unstructured.Unstructured{}

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow has no pipeline spec annotation is invalid", func() {
			It("Errors", func() {
				workflow := &unstructured.Unstructured{}
				workflow.SetAnnotations(map[string]string{
					pipelineSpecAnnotationName: fmt.Sprintf(`{invalid`),
				})

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow has no pipeline spec annotation is empty", func() {
			It("Errors", func() {
				workflow := &unstructured.Unstructured{}
				workflow.SetAnnotations(map[string]string{
					pipelineSpecAnnotationName: fmt.Sprintf(`{"name": ""}`),
				})

				_, err := getPipelineName(workflow)
				Expect(err).To(HaveOccurred())
			})
		})

		When("The workflow has no pipeline spec annotation is missing", func() {
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
