package workflows

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	When("getWorkflowOutput is called with a worklow that has a output with the given key", func() {
		It("returs the output value", func() {
			workflow := argo.Workflow{
				Spec: argo.WorkflowSpec{
					Entrypoint: "entrypoint",
				},
				Status: argo.WorkflowStatus{
					Nodes: map[string]argo.NodeStatus{
						"entrypoint": argo.NodeStatus{
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
			result, error := GetWorkflowOutput(&workflow, "aKey")
			Expect(error).NotTo(HaveOccurred())
			Expect(result).To(Equal("aValue"))
		})
	})
})
