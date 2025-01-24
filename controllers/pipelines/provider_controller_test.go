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

	var _ = Describe("syncDeployment", func() {
		Specify("Should update the existing Deployment to match the desired Deployment, but leave the Status unchanged", func() {
			existingDeployment := appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"replace-me": "this-label-has-no-place-here",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "container",
									Image: "unwanted-image",
								},
							},
						},
					},
				},
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:   appsv1.DeploymentProgressing,
							Status: v1.ConditionTrue,
						},
						{
							Type:   appsv1.DeploymentAvailable,
							Status: v1.ConditionTrue,
						},
					},
					CollisionCount: new(int32),
				},
			}

			desiredDeployment := appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"proper-label": "this-is-the-right-label",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "container",
									Image: "correct-image",
								},
							},
						},
					},
				},
				Status: appsv1.DeploymentStatus{},
			}
			setResourceHashAnnotation(&desiredDeployment)

			syncedDeployment := syncDeployment(&existingDeployment, &desiredDeployment)

			Expect(syncedDeployment.Spec).To(Equal(desiredDeployment.Spec))
			Expect(syncedDeployment.Labels).To(Equal(desiredDeployment.Labels))
			Expect(syncedDeployment.Annotations[ResourceHashAnnotation]).To(Equal(desiredDeployment.Annotations[ResourceHashAnnotation]))
			Expect(syncedDeployment.Status).To(Equal(existingDeployment.Status))
		})
	})
})
