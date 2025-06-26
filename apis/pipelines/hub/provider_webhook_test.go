package v1beta1

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common/testutil/mocks"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ProviderValidator Webhook", func() {
	var (
		mockReader mocks.MockK8sClientReader
		validator  ProviderValidator
		provider   Provider
		pipeline   Pipeline
		ctx        = context.Background()
	)

	BeforeEach(func() {
		mockReader = mocks.MockK8sClientReader{}
		provider = Provider{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "provider-namespace",
				Name:      "provider-name",
			},
			Spec: ProviderSpec{
				Frameworks: []Framework{
					{Name: "framework-name-1"},
					{Name: "framework-name-2"},
				},
				AllowedNamespaces: []string{"ns-1"},
			},
		}
		pipeline = Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns-1",
				Name:      "pipeline-name",
			},
			Spec: PipelineSpec{
				Provider: provider.GetCommonNamespacedName(),
				Framework: PipelineFramework{
					Name: "framework-name-1",
				},
			},
		}
		validator = ProviderValidator{
			reader: &mockReader,
		}
	})

	Context("validateDelete", func() {
		When("the runtime object is not a provider", func() {
			It("should return a StatusError", func() {
				warnings, err := validator.ValidateDelete(ctx, &Run{})
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonBadRequest))
			})
		})

		When("k8s reader errors when fetching the pipeline list", func() {
			It("should return a StatusError", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(errors.New("No pipeline list"))

				warnings, err := validator.ValidateDelete(ctx, &provider)
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonInternalError))
			})
		})

		When("k8s reader returns pipelines that references provider", func() {
			It("should return a StatusError", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(
					func(args mock.Arguments) {
						pipelineList := args.Get(0).(*PipelineList)
						nonMatchingPipeline := pipeline
						nonMatchingPipeline.Spec.Framework.Name = "other-framework"
						matchingPipeline1 := pipeline
						matchingPipeline2 := pipeline
						matchingPipeline2.Spec.Framework.Name = "framework-name-2"

						pipelineList.Items = append(
							pipelineList.Items,
							nonMatchingPipeline,
							matchingPipeline1,
							matchingPipeline2,
						)
					},
				)

				warnings, err := validator.ValidateDelete(ctx, &provider)
				Expect(warnings).To(BeNil())
				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Details.Causes[0].Type).To(Equal(metav1.CauseTypeForbidden))
				Expect(statusErr.Status().Details.Causes[1].Type).To(Equal(metav1.CauseTypeForbidden))
			})
		})

		When("k8s reader returns pipelines that do not reference the provider", func() {
			It("should not error", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(
					func(args mock.Arguments) {
						pipelineList := args.Get(0).(*PipelineList)
						nonMatchingPipeline1 := pipeline
						nonMatchingPipeline1.Spec.Framework.Name = "other-framework-1"
						nonMatchingPipeline2 := pipeline
						nonMatchingPipeline2.Spec.Framework.Name = "other-framework-2"
						pipelineList.Items = append(
							pipelineList.Items,
							nonMatchingPipeline1,
							nonMatchingPipeline2,
						)
					},
				)

				warnings, err := validator.ValidateDelete(ctx, &provider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("k8s reader returns no pipelines", func() {
			It("should not error", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {})

				warnings, err := validator.ValidateDelete(ctx, &provider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("validateUpdate", func() {

		newProvider := provider.DeepCopy()
		oldProvider := provider.DeepCopy()
		frameworkMarkedForDeletion := "framework-which-would-be-deleted-by-update"
		oldProvider.Spec.Frameworks = append(oldProvider.Spec.Frameworks, Framework{Name: frameworkMarkedForDeletion})

		When("the new runtime object is not a provider", func() {
			It("should return a StatusError", func() {
				warnings, err := validator.ValidateUpdate(ctx, &Provider{}, &Run{})
				Expect(warnings).To(BeNil())

				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonBadRequest))
			})
		})

		When("the old runtime object is not a provider", func() {
			It("should return a StatusError", func() {
				warnings, err := validator.ValidateUpdate(ctx, &Run{}, &Provider{})
				Expect(warnings).To(BeNil())

				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonBadRequest))
			})
		})

		When("k8s reader errors when fetching the pipeline list", func() {
			It("should return a StatusError", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(errors.New("No pipeline list"))

				warnings, err := validator.ValidateUpdate(ctx, oldProvider, newProvider)
				Expect(warnings).To(BeNil())

				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonInternalError))
			})
		})

		When("k8s reader returns pipelines that references provider with a framework marked for deletion", func() {
			It("should return a StatusError", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(
					func(args mock.Arguments) {
						pipelineList := args.Get(0).(*PipelineList)
						nonMatchingPipeline := pipeline
						matchingPipeline := pipeline
						matchingPipeline.Spec.Framework.Name = frameworkMarkedForDeletion

						pipelineList.Items = append(
							pipelineList.Items,
							nonMatchingPipeline,
							matchingPipeline,
						)
					},
				)

				warnings, err := validator.ValidateUpdate(ctx, oldProvider, newProvider)
				Expect(warnings).To(BeNil())

				var statusErr *apierrors.StatusError
				Expect(errors.As(err, &statusErr)).To(BeTrue())
				Expect(statusErr.Status().Details.Causes[0].Type).To(Equal(metav1.CauseTypeForbidden))
			})
		})

		When("k8s reader returns pipelines that do not reference a providers framework marked for deletion", func() {
			It("should not error", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(
					func(args mock.Arguments) {
						pipelineList := args.Get(0).(*PipelineList)
						nonMatchingPipeline1 := pipeline
						nonMatchingPipeline1.Spec.Framework.Name = "other-framework-1"
						nonMatchingPipeline2 := pipeline
						nonMatchingPipeline2.Spec.Framework.Name = "other-framework-2"
						pipelineList.Items = append(
							pipelineList.Items,
							nonMatchingPipeline1,
							nonMatchingPipeline2,
						)
					},
				)

				warnings, err := validator.ValidateUpdate(ctx, oldProvider, newProvider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("k8s reader returns pipelines that do not reference the provider", func() {
			It("should not error", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(
					func(args mock.Arguments) {
						pipelineList := args.Get(0).(*PipelineList)
						matchingPipeline := pipeline

						pipelineList.Items = append(
							pipelineList.Items,
							matchingPipeline,
						)
					},
				)

				warnings, err := validator.ValidateUpdate(ctx, oldProvider, newProvider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("k8s reader returns no pipelines", func() {
			It("should not error", func() {
				mockReader.On(
					"List",
					mock.AnythingOfType("*v1beta1.PipelineList"),
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {})

				warnings, err := validator.ValidateUpdate(ctx, oldProvider, newProvider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

})
