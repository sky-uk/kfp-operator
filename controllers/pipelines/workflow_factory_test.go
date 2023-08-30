//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

var _ = Describe("CommonWorkflowMeta", func() {
	It("creates metadata", func() {
		owner := pipelinesv1.RandomResource()
		namespace := RandomString()
		w := ResourceWorkflowFactory[*pipelinesv1.TestResource, interface{}]{
			Config: config.Configuration{
				WorkflowNamespace: namespace,
			},
		}
		meta := w.CommonWorkflowMeta(owner)

		Expect(meta.Namespace).To(Equal(namespace))
		Expect(meta.GetGenerateName()).To(Equal(owner.GetKind() + "-" + owner.GetName() + "-"))

		Expect(meta.Labels[WorkflowConstants.OwnerKindLabelKey]).To(Equal(owner.GetKind()))
		Expect(meta.Labels[WorkflowConstants.OwnerNameLabelKey]).To(Equal(owner.GetName()))
		Expect(meta.Labels[WorkflowConstants.OwnerNamespaceLabelKey]).To(Equal(owner.GetNamespace()))
	})
})
