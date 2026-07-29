//go:build unit

package v1beta1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ProviderValidator Webhook", func() {
	var (
		validator ProviderValidator
		provider  Provider
	)

	BeforeEach(func() {
		validator = ProviderValidator{}
		provider = Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provider-name",
				Namespace: "provider-ns",
			},
		}
	})

	Context("validate", func() {
		When("podTemplateEnv contains no reserved names", func() {
			It("should not error", func() {
				provider.Spec.PodTemplateEnv = []corev1.EnvVar{
					{Name: "KUBE_FEATURE_WatchListClient", Value: "true"},
					{Name: "HTTPS_PROXY", Value: "http://proxy:3128"},
				}

				warnings, err := validator.validate(&provider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("podTemplateEnv is empty", func() {
			It("should not error", func() {
				warnings, err := validator.validate(&provider)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		DescribeTable("rejecting reserved env var names",
			func(name string) {
				provider.Spec.PodTemplateEnv = []corev1.EnvVar{
					{Name: name, Value: "x"},
				}

				warnings, err := validator.validate(&provider)
				Expect(warnings).To(BeNil())

				statusErr := &apierrors.StatusError{}
				Expect(apierrors.IsInvalid(err)).To(BeTrue())
				Expect(err).To(BeAssignableToTypeOf(statusErr))
				causes := err.(*apierrors.StatusError).Status().Details.Causes
				Expect(causes).To(HaveLen(1))
				Expect(causes[0].Type).To(Equal(metav1.CauseTypeForbidden))
			},
			Entry("PROVIDERNAME", "PROVIDERNAME"),
			Entry("PIPELINEROOTSTORAGE", "PIPELINEROOTSTORAGE"),
			Entry("PARAMETERS_ prefix", "PARAMETERS_FOO"),
			Entry("lowercase providername", "providername"),
			Entry("lowercase parameters prefix", "parameters_foo"),
		)

		When("multiple reserved names are present", func() {
			It("should report every offending entry", func() {
				provider.Spec.PodTemplateEnv = []corev1.EnvVar{
					{Name: "SAFE", Value: "ok"},
					{Name: "PROVIDERNAME", Value: "x"},
					{Name: "PARAMETERS_KEY", Value: "y"},
				}

				warnings, err := validator.validate(&provider)
				Expect(warnings).To(BeNil())
				Expect(apierrors.IsInvalid(err)).To(BeTrue())
				causes := err.(*apierrors.StatusError).Status().Details.Causes
				Expect(causes).To(HaveLen(2))
			})
		})
	})
})
