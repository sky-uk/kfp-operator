//go:build unit

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Utils", func() {
	When("getWorkflowOutput is called with a workflow that has an output with the given key", func() {
		It("returns the output value", func() {
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
										Value: argo.AnyStringPtr("id: aValue"),
									},
								},
							},
						},
					},
				},
			}
			result, error := getWorkflowOutput(&workflow, "aKey")
			Expect(error).NotTo(HaveOccurred())
			Expect(result.Id).To(Equal("aValue"))
		})
	})

	When("latestWorkflowByPhase is called with one workflow each", func() {
		It("returns all values", func() {
			expectedInProgress := argo.Workflow{
				Status: argo.WorkflowStatus{
					Phase: argo.WorkflowPending,
				},
			}

			expectedFailed := argo.Workflow{
				Status: argo.WorkflowStatus{
					Phase: argo.WorkflowFailed,
				},
			}

			expectedSucceeded := argo.Workflow{
				Status: argo.WorkflowStatus{
					Phase: argo.WorkflowSucceeded,
				},
			}

			inProgress, succeeded, failed := latestWorkflowByPhase([]argo.Workflow{
				expectedInProgress, expectedSucceeded, expectedFailed,
			})

			Expect(inProgress).To(Equal(&expectedInProgress))
			Expect(succeeded).To(Equal(&expectedSucceeded))
			Expect(failed).To(Equal(&expectedFailed))
		})
	})

	When("latestWorkflowByPhase is called with multiple per phase", func() {
		It("returns the latest", func() {
			expectedSucceededOldest := argo.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Unix(0, 0),
				},
				Status: argo.WorkflowStatus{
					Phase: argo.WorkflowSucceeded,
				},
			}

			expectedSucceededNewest := argo.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Unix(1, 0),
				},
				Status: argo.WorkflowStatus{
					Phase: argo.WorkflowSucceeded,
				},
			}

			_, succeeded, _ := latestWorkflowByPhase([]argo.Workflow{
				expectedSucceededNewest, expectedSucceededOldest,
			})

			Expect(succeeded).To(Equal(&expectedSucceededNewest))
		})
	})

	When("latestWorkflowByPhase is called with no workflows", func() {
		It("returns all empty values", func() {
			inProgress, succeeded, failed := latestWorkflowByPhase([]argo.Workflow{})
			Expect(inProgress).To(BeNil())
			Expect(succeeded).To(BeNil())
			Expect(failed).To(BeNil())
		})
	})
})
