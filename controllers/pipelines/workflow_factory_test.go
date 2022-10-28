//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Describe("CommonWorkflowMeta", func() {
	It("creates metadata", func() {
		owner := pipelinesv1.RandomResource()
		operation := RandomString()

		meta := CommonWorkflowMeta(owner, operation)

		Expect(meta.Namespace).To(Equal(owner.GetNamespace()))
		Expect(meta.GetGenerateName()).To(Equal(operation + "-" + owner.GetKind() + "-"))

		Expect(meta.Labels[WorkflowConstants.OwnerKindLabelKey]).To(Equal(owner.GetKind()))
		Expect(meta.Labels[WorkflowConstants.OwnerNameLabelKey]).To(Equal(owner.GetName()))
		Expect(meta.Labels[WorkflowConstants.OperationLabelKey]).To(Equal(operation))
	})
})
