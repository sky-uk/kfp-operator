//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("Provider Controller", func() {
	var _ = Describe("constructDeployment", func() {
		Specify("Should construct a Deployment with correct values given config and a Provider resource", func() {
			config := config.KfpControllerConfigSpec{
				DefaultProviderValues: config.DefaultProviderValues{
					Replicas:             2,
					ServiceContainerName: "container-name",
					PodTemplateSpec: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "container-name",
								},
							},
						},
					},
				},
			}

			provider := pipelinesv1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-provider",
					Namespace: "my-ns",
				},
				Spec: pipelinesv1.ProviderSpec{
					ServiceImage: "image",
				},
			}

			replicas := int32(2)
			expectedDeployment := appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "provider-my-provider-",
					Namespace:    provider.Namespace,
					Labels: map[string]string{
						"app": "provider-my-provider",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "provider-my-provider",
						},
					},
					Replicas: &replicas,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "provider-my-provider",
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "container-name",
									Image: "image",
									Env: []v1.EnvVar{
										{
											Name:  "PROVIDERNAME",
											Value: "my-provider",
										},
									},
								},
							},
						},
					},
				},
			}
			setResourceHashAnnotation(&expectedDeployment)

			actualDeployment, err := constructDeployment(&provider, config)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualDeployment.Spec.Selector).To(Equal(expectedDeployment.Spec.Selector))
			Expect(actualDeployment.Spec.Replicas).To(Equal(expectedDeployment.Spec.Replicas))
			Expect(actualDeployment.Spec.Template.ObjectMeta).To(Equal(expectedDeployment.Spec.Template.ObjectMeta))
			Expect(actualDeployment.Spec.Template.Spec.Containers).To(Equal(expectedDeployment.Spec.Template.Spec.Containers))
			Expect(actualDeployment.ObjectMeta).To(Equal(expectedDeployment.ObjectMeta))
		})
	})
})
