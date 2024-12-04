//go:build unit

package sources

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	workflowPhaseLabel = "workflows.argoproj.io/phase"
)

func setWorkflowPhase(workflow *unstructured.Unstructured, phase argo.WorkflowPhase) {
	workflow.SetLabels(map[string]string{
		workflowPhaseLabel: string(phase),
	})
}

func runCompletionStatus(workflow *unstructured.Unstructured) (common.RunCompletionStatus, bool) {
	switch workflow.GetLabels()[workflowPhaseLabel] {
	case string(argo.WorkflowSucceeded):
		return common.RunCompletionStatuses.Succeeded, true
	case string(argo.WorkflowFailed), string(argo.WorkflowError):
		return common.RunCompletionStatuses.Failed, true
	default:
		return "", false
	}
}

var _ = Context("workflow source", func() {
	Describe("jsonPatchPath", func() {
		It("concatenates path segments", func() {
			segment1 := common.RandomString()
			segment2 := common.RandomString()
			segment3 := common.RandomString()

			expectedPath := fmt.Sprintf("/%s/%s/%s", segment1, segment2, segment3)

			Expect(jsonPatchPath(segment1, segment2, segment3)).To(Equal(expectedPath))
		})

		It("escapes '/'", func() {
			segment1 := common.RandomString()
			segment2 := common.RandomString()

			toBeEscaped := fmt.Sprintf("%s/%s", segment1, segment2)
			escaped := fmt.Sprintf("/%s~1%s", segment1, segment2)
			Expect(jsonPatchPath(toBeEscaped)).To(Equal(escaped))
		})

		It("escapes '~'", func() {
			segment1 := common.RandomString()
			segment2 := common.RandomString()

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
		func(phase argo.WorkflowPhase, expectedStatus common.RunCompletionStatus) {
			workflow := &unstructured.Unstructured{}
			setWorkflowPhase(workflow, phase)
			status, hasFinished := runCompletionStatus(workflow)
			Expect(status).To(Equal(expectedStatus))
			Expect(hasFinished).To(Equal(true))
		},
		Entry("error", argo.WorkflowError, common.RunCompletionStatuses.Failed),
		Entry("failed", argo.WorkflowFailed, common.RunCompletionStatuses.Failed),
		Entry("succeeded", argo.WorkflowSucceeded, common.RunCompletionStatuses.Succeeded),
	)
})
