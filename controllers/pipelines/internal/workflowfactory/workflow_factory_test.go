//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
)

var _ = Describe("CommonWorkflowMeta", func() {
	It("creates metadata", func() {
		owner := pipelinesv1.RandomResource()
		namespace := RandomString()
		w := ResourceWorkflowFactory[*pipelinesv1.TestResource, interface{}]{
			Config: config.KfpControllerConfigSpec{
				WorkflowNamespace: namespace,
			},
		}
		meta := w.CommonWorkflowMeta(owner)

		Expect(meta.Namespace).To(Equal(namespace))
		Expect(meta.GetGenerateName()).To(Equal(owner.GetKind() + "-" + owner.GetName() + "-"))

		Expect(meta.Labels[workflowconstants.OwnerKindLabelKey]).To(Equal(owner.GetKind()))
		Expect(meta.Labels[workflowconstants.OwnerNameLabelKey]).To(Equal(owner.GetName()))
		Expect(meta.Labels[workflowconstants.OwnerNamespaceLabelKey]).To(Equal(owner.GetNamespace()))
	})
})
