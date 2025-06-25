//go:build unit

package v1beta1

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/common/testutil/mocks"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PipelineValidator Webhook", func() {
	var (
		mockReader mocks.MockK8sClientReader
		validator  PipelineValidator
		pipeline   Pipeline
		ctx        = context.Background()
	)

	BeforeEach(func() {
		mockReader = mocks.MockK8sClientReader{}
		pipeline = Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "pipeline-ns",
			},
			Spec: PipelineSpec{
				Provider: common.NamespacedName{
					Name:      "provider-name",
					Namespace: "provider-ns",
				},
				Framework: PipelineFramework{
					Name: "framework-name",
				},
			},
		}
		validator = PipelineValidator{
			reader: &mockReader,
		}
	})

	Context("validate", func() {
		When("k8s reader returns specified provider and it contains a matching framework", func() {
			It("not error", func() {
				mockReader.On(
					"Get",
					client.ObjectKey{
						Namespace: pipeline.Spec.Provider.Namespace,
						Name:      pipeline.Spec.Provider.Name,
					},
					mock.AnythingOfType("*v1beta1.Provider"),
				).Return(nil).Run(
					func(args mock.Arguments) {
						provider := args.Get(1).(*Provider)
						provider.Spec.AllowedNamespaces = []string{pipeline.GetNamespace()}
						provider.Spec.Frameworks = []Framework{
							{Name: "some-other-framework"},
							{Name: pipeline.Spec.Framework.Name},
						}
					},
				)
				warnings, err := validator.validate(ctx, &pipeline)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("k8s reader returns a provider that does not contain the specified framework", func() {
			It("should return a StatusError", func() {
				mockReader.On(
					"Get",
					client.ObjectKey{
						Namespace: pipeline.Spec.Provider.Namespace,
						Name:      pipeline.Spec.Provider.Name,
					},
					mock.AnythingOfType("*v1beta1.Provider"),
				).Return(nil).Run(
					func(args mock.Arguments) {
						provider := args.Get(1).(*Provider)
						provider.Spec.AllowedNamespaces = []string{pipeline.GetNamespace()}
						provider.Spec.Frameworks = []Framework{
							{Name: "some-other-framework"},
							{Name: "another-frame-work"},
						}
					},
				)

				warnings, err := validator.validate(ctx, &pipeline)
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Details.Causes[0].Type).To(Equal(metav1.CauseTypeFieldValueNotSupported))
			})
		})

		When("k8s reader errors when fetching the specified provider", func() {
			It("should return an StatusError", func() {
				mockReader.On(
					"Get",
					client.ObjectKey{
						Namespace: pipeline.Spec.Provider.Namespace,
						Name:      pipeline.Spec.Provider.Name,
					},
					mock.AnythingOfType("*v1beta1.Provider"),
				).Return(errors.New("No matching provider"))

				warnings, err := validator.validate(ctx, &pipeline)
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Details.Causes[0].Type).To(Equal(metav1.CauseTypeFieldValueNotFound))
			})
		})

		When("the pipeline namespace is not allowed by the provider", func() {
			It("should return a StatusError", func() {
				mockReader.On(
					"Get",
					client.ObjectKey{
						Namespace: pipeline.Spec.Provider.Namespace,
						Name:      pipeline.Spec.Provider.Name,
					},
					mock.AnythingOfType("*v1beta1.Provider"),
				).Return(nil).Run(
					func(args mock.Arguments) {
						provider := args.Get(1).(*Provider)
						provider.Spec.AllowedNamespaces = []string{"some-other-namespace"}
					},
				)

				warnings, err := validator.validate(ctx, &pipeline)
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Details.Causes[0].Type).To(Equal(metav1.CauseTypeForbidden))
			})
		})

		When("the provider allows all namespaces", func() {
			It("should not error", func() {
				mockReader.On(
					"Get",
					client.ObjectKey{
						Namespace: pipeline.Spec.Provider.Namespace,
						Name:      pipeline.Spec.Provider.Name,
					},
					mock.AnythingOfType("*v1beta1.Provider"),
				).Return(nil).Run(
					func(args mock.Arguments) {
						provider := args.Get(1).(*Provider)
						provider.Spec.Frameworks = []Framework{
							{Name: "some-other-framework"},
							{Name: pipeline.Spec.Framework.Name},
						}
					},
				)

				warnings, err := validator.validate(ctx, &pipeline)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("the runtime object is not a pipeline", func() {
			It("should return a StatusError", func() {
				warnings, err := validator.validate(ctx, &Run{})
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonBadRequest))
			})
		})
	})
})
