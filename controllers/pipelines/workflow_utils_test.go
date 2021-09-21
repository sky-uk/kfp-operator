// +build unit

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Utils", func() {
	When("getWorkflowOutput is called with a worklow that has a output with the given key", func() {
		It("returs the output value", func() {
			workflow := argo.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "a-workflow",
				},
				Status: argo.WorkflowStatus{
					Nodes: map[string]argo.NodeStatus{
						"a-workflow": {
							Outputs: &argo.Outputs{
								Parameters: []argo.Parameter{
									{
										Name:  "aKey",
										Value: argo.AnyStringPtr("aValue"),
									},
								},
							},
						},
					},
				},
			}
			result, error := getWorkflowOutput(&workflow, "aKey")
			Expect(error).NotTo(HaveOccurred())
			Expect(result).To(Equal("aValue"))
		})
	})
})
