//go:build unit

package pipelines

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8Scheme "k8s.io/client-go/kubernetes/scheme"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Context("Provider Service Manager", func() {

	var (
		ctx            = context.Background()
		provider       *pipelineshub.Provider
		client         k8sClient.Client
		serviceManager ServiceManager
	)

	BeforeEach(func() {
		provider = pipelineshub.RandomProvider()
		client = fake.NewClientBuilder().
			WithScheme(testutil.SchemeWithCrds()).
			WithScheme(k8Scheme.Scheme).
			Build()

		optInClient := &controllers.OptInClient{
			Writer:       client,
			StatusClient: client,
			Cached:       client,
			NonCached:    client,
		}

		serviceManager = ServiceManager{
			client: optInClient,
			scheme: client.Scheme(),
			config: &configv1.KfpControllerConfigSpec{
				DefaultProviderValues: configv1.DefaultProviderValues{
					ServiceContainerName: "ServiceContainerName",
					ServicePort:          9999,
				},
			},
		}
	})

	var _ = Describe("Create", func() {

		It("should not error if the service is created successfully", func() {
			service := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			}

			err := serviceManager.Create(ctx, &service, provider)
			Expect(err).ToNot(HaveOccurred())

			result := &corev1.Service{}
			err = client.Get(ctx, k8sClient.ObjectKey{Name: provider.Name, Namespace: provider.Namespace}, result)
			Expect(err).ToNot(HaveOccurred())

			Expect(result.Name).To(Equal(service.Name))
			Expect(result.Namespace).To(Equal(service.Namespace))
		})

		It("should return an error if the service is not created successfully", func() {
			service := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			}

			// Create an existing service to trigger error on attempting to create another
			err := client.Create(ctx, &service)
			Expect(err).ToNot(HaveOccurred())

			err = serviceManager.Create(ctx, &service, provider)
			Expect(err).To(HaveOccurred())
		})
	})

	var _ = Describe("Delete", func() {

		It("should not error if the service is deleted successfully", func() {
			service := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			}

			err := client.Create(ctx, &service)
			Expect(err).ToNot(HaveOccurred())

			serviceList := &corev1.ServiceList{}
			err = client.List(ctx, serviceList)
			Expect(err).ToNot(HaveOccurred())
			Expect(serviceList.Items).To(HaveLen(1))

			err = serviceManager.Delete(ctx, &service)
			Expect(err).ToNot(HaveOccurred())

			err = client.List(ctx, serviceList)
			Expect(err).ToNot(HaveOccurred())
			Expect(serviceList.Items).To(BeEmpty())
		})

		It("should return an error if the service is not deleted successfully", func() {
			service := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			}

			// delete something that does not exist within k8s
			err := serviceManager.Delete(ctx, &service)
			Expect(err).To(HaveOccurred())
		})
	})

	var _ = Describe("Get", func() {

		It("should return the service owned by a provider if it exists", func() {
			service := corev1.Service{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.Name,
					Namespace: provider.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(provider, pipelineshub.GroupVersion.WithKind("Provider")),
					},
				},
			}

			err := client.Create(ctx, &service)
			Expect(err).ToNot(HaveOccurred())

			result, err := serviceManager.Get(ctx, provider)
			Expect(err).ToNot(HaveOccurred())

			Expect(result.Name).To(Equal(service.Name))
			Expect(result.Namespace).To(Equal(service.Namespace))
		})

		It("should return a NotFound error if the service does not exist", func() {
			service, err := serviceManager.Get(ctx, provider)
			Expect(service).To(BeNil())
			Expect(err).To(Equal(apierrors.NewNotFound(schema.GroupResource{}, "")))
		})
	})

	var _ = Describe("Construct", func() {

		It("should return a service with expected values from provider", func() {
			result := serviceManager.Construct(provider)

			providerSuffixedName := fmt.Sprintf("provider-%s", provider.Name)

			expectedPorts := []corev1.ServicePort{
				{
					Name:       "http",
					Port:       int32(serviceManager.config.DefaultProviderValues.ServicePort),
					TargetPort: intstr.FromInt(serviceManager.config.DefaultProviderValues.ServicePort),
					Protocol:   corev1.ProtocolTCP,
				},
			}

			Expect(result.GenerateName).To(Equal(providerSuffixedName + "-"))
			Expect(result.Namespace).To(Equal(provider.Namespace))
			Expect(result.Spec.Ports).To(Equal(expectedPorts))
			Expect(result.Spec.Selector).To(Equal(map[string]string{"app": providerSuffixedName}))
		})
	})

	var _ = Describe("Equal", func() {
		It("should return true if the services are equal", func() {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      provider.Name,
					Namespace: provider.Namespace,
				},
			}

			Expect(serviceManager.Equal(service, service)).To(BeTrue())
		})

		It("should return false if the services are not equal", func() {
			service := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{
						Port: 80,
					}},
				},
			}

			service2 := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{
						Port: 9999,
					}},
				},
			}
			Expect(serviceManager.Equal(service, service2)).To(BeFalse())
		})
	})
})
