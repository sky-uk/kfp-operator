//go:build unit

package pipelines

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8Scheme "k8s.io/client-go/kubernetes/scheme"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Context("Provider Deployment Manager", func() {

	var (
		ctx               = context.Background()
		provider          = pipelineshub.RandomProvider()
		client            k8sClient.Client
		deploymentManager DeploymentManager
	)

	BeforeEach(func() {
		client = fake.NewClientBuilder().
			WithScheme(testutil.SchemeWithCrds()).
			WithScheme(k8Scheme.Scheme).
			WithObjects(&appsv1.Deployment{}).
			Build()

		optInClient := &controllers.OptInClient{
			Writer:       client,
			StatusClient: client,
			Cached:       client,
			NonCached:    client,
		}

		deploymentManager = DeploymentManager{
			client: optInClient,
			scheme: client.Scheme(),
			config: &configv1.KfpControllerConfigSpec{
				DefaultProviderValues: configv1.DefaultProviderValues{
					ServiceContainerName: "ServiceContainerName",
					PodTemplateSpec: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "ServiceContainerName"},
							},
						},
					},
				},
			},
		}
	})

	var _ = Describe("Create", func() {

		Specify("should not error if the deployment is created successfully", func() {
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
			}

			err := deploymentManager.Create(context.Background(), &deployment, provider)
			Expect(err).ToNot(HaveOccurred())

			result := &appsv1.Deployment{}
			err = client.Get(ctx, k8sClient.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, result)
			Expect(err).ToNot(HaveOccurred())

			Expect(result.Name).To(Equal(deployment.Name))
			Expect(result.Namespace).To(Equal(deployment.Namespace))
		})

		Specify("Should return an error if the client create fails", func() {
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
			}

			err := client.Create(ctx, &deployment)
			Expect(err).ToNot(HaveOccurred())

			// attempting to create an already existing deployment will trigger error
			err = deploymentManager.Create(context.Background(), &deployment, provider)
			Expect(err).To(HaveOccurred())
		})

	})

	var _ = Describe("Update", func() {

		Specify("should not error if the deployment is updated successfully", func() {
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
			}

			err := client.Create(ctx, &deployment)
			Expect(err).ToNot(HaveOccurred())

			replicas := int32(2)
			deployment.Spec.Replicas = &replicas

			err = deploymentManager.Update(context.Background(), &deployment, &deployment, provider)
			Expect(err).ToNot(HaveOccurred())

			result := &appsv1.Deployment{}
			err = client.Get(ctx, k8sClient.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, result)
			Expect(err).ToNot(HaveOccurred())

			// match up the modified field
			Expect(result.Spec.Replicas).To(Equal(&replicas))
		})

		Specify("Should return an error if the deployment update fails", func() {
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
			}

			// attempting to update a non-existing deployment will trigger error
			err := deploymentManager.Update(context.Background(), &deployment, &deployment, provider)
			Expect(err).To(HaveOccurred())
		})

	})

	var _ = Describe("Get", func() {

		It("should return the deployment owned by a provider if it exists", func() {
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(provider, pipelineshub.GroupVersion.WithKind("Provider")),
					},
				},
			}

			err := client.Create(ctx, &deployment)
			Expect(err).ToNot(HaveOccurred())

			result, err := deploymentManager.Get(context.Background(), provider)
			Expect(err).ToNot(HaveOccurred())

			Expect(result.Name).To(Equal(deployment.Name))
			Expect(result.Namespace).To(Equal(deployment.Namespace))
		})

		It("should return a NotFound error if the deployment does not exist", func() {
			result, err := deploymentManager.Get(context.Background(), provider)
			Expect(err).To(Equal(apierrors.NewNotFound(schema.GroupResource{}, "")))
			Expect(result).To(BeNil())
		})

	})

	var _ = Describe("Construct", func() {

		Specify("Should return the constructed deployment if it is successful", func() {
			provider.Spec.Parameters = map[string]*apiextensionsv1.JSON{
				"key1": {Raw: []byte(`"value1"`)},
				"key2": {Raw: []byte(`1`)},
			}

			deployment, err := deploymentManager.Construct(provider)
			Expect(err).ToNot(HaveOccurred())

			providerSuffixedName := fmt.Sprintf("provider-%s", provider.Name)

			Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(Equal([]corev1.EnvVar{
				{
					Name:  "PARAMETERS_KEY1",
					Value: "value1",
				},
				{
					Name:  "PARAMETERS_KEY2",
					Value: "1",
				},
				{
					Name:  ProviderNameEnvVar,
					Value: provider.Name,
				},
			}))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(provider.Spec.ServiceImage))
			Expect(deployment.Spec.Template.Labels).To(Equal(map[string]string{
				"app": providerSuffixedName,
			}))
			Expect(deployment.GenerateName).To(Equal(providerSuffixedName + "-"))
			Expect(deployment.Namespace).To(Equal(provider.Namespace))
		})

		Specify("Should return an error if the no container with matching ServiceContainerName exists", func() {
			deploymentManager.config.DefaultProviderValues.PodTemplateSpec.Spec.Containers = []corev1.Container{}
			_, err := deploymentManager.Construct(provider)

			Expect(err).To(HaveOccurred())
		})

	})

	var _ = Describe("Equal", func() {

		Specify("Should return true if the deployments are equal", func() {
			deployment1 := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "ServiceContainerName"}}},
					},
				},
			}

			deployment2 := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "ServiceContainerName"}}},
					},
				},
			}

			result := deploymentManager.Equal(&deployment1, &deployment2)
			Expect(result).To(BeTrue())
		})

		Specify("Should return false if the deployments are not equal", func() {
			deployment1 := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "ServiceContainerName"}}},
					},
				},
			}

			deployment2 := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "NonMatchingContainerName"}}},
					},
				},
			}

			result := deploymentManager.Equal(&deployment1, &deployment2)
			Expect(result).To(BeFalse())
		})
	})

	var _ = Describe("jsonToString", func() {
		Specify("Should return a plain string given a JSON string (no extra quotes or escape chars!)", func() {
			rawJson, err := json.Marshal("test")
			Expect(err).ToNot(HaveOccurred())

			jsonStr := apiextensionsv1.JSON{
				Raw: rawJson,
			}

			result := jsonToString(&jsonStr)

			Expect(result).To(Equal("test"))
		})

		Specify("Should return a raw JSON string given a JSON object", func() {
			rawJson, err := json.Marshal(`{"key1": "value1", "key2": 42, "key3": {"key4": "value4"}, "key5": ""}`)
			Expect(err).ToNot(HaveOccurred())

			jsonStr := apiextensionsv1.JSON{
				Raw: rawJson,
			}

			result := jsonToString(&jsonStr)

			Expect(result).To(Equal(`{"key1": "value1", "key2": 42, "key3": {"key4": "value4"}, "key5": ""}`))
		})
	})

})
