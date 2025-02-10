package pipelines

import (
	"context"
	"encoding/json"
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	. "github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/predicates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

type DeploymentResourceManager interface {
	Create(ctx context.Context, new *appsv1.Deployment, owner *pipelinesv1.Provider) error
	Update(ctx context.Context, old, new *appsv1.Deployment, owner *pipelinesv1.Provider) error
	Get(ctx context.Context, owner *pipelinesv1.Provider) (*appsv1.Deployment, error)
	Equal(a, b *appsv1.Deployment) bool
	Construct(provider *pipelinesv1.Provider) (*appsv1.Deployment, error)
}

type DeploymentManager struct {
	client *controllers.OptInClient
	scheme *runtime.Scheme
	config *config.KfpControllerConfigSpec
}

func (dm DeploymentManager) Create(ctx context.Context, new *appsv1.Deployment, owner *pipelinesv1.Provider) error {
	logger := log.FromContext(ctx)

	if err := ctrl.SetControllerReference(owner, new, dm.scheme); err != nil {
		logger.Error(err, "unable to set controller reference on deployment")
		return err
	}

	if new.Annotations == nil {
		new.Annotations = make(map[string]string)
	}
	new.Annotations[predicates.ControllerManagedKey] = "true"
	if err := dm.client.Create(ctx, new); err != nil {
		logger.Error(err, "unable to create provider deployment")
		return err
	}
	return nil
}

func (dm DeploymentManager) Update(ctx context.Context, old *appsv1.Deployment, new *appsv1.Deployment, owner *pipelinesv1.Provider) error {
	logger := log.FromContext(ctx)

	old.Spec = new.Spec
	old.SetLabels(new.Labels)

	if err := ctrl.SetControllerReference(owner, old, dm.scheme); err != nil {
		logger.Error(err, "unable to set controller reference on deployment")
		return err
	}

	if err := dm.client.Update(ctx, old); err != nil {
		logger.Error(err, "unable to update provider deployment", "deployment", new)
		return err
	}
	return nil
}

func (dm DeploymentManager) Get(ctx context.Context, owner *pipelinesv1.Provider) (*appsv1.Deployment, error) {
	dl := &appsv1.DeploymentList{}
	if err := dm.client.NonCached.List(ctx, dl, &client.ListOptions{
		Namespace: owner.Namespace,
	}); err != nil {
		return nil, err
	}

	for _, deploy := range dl.Items {
		if metav1.IsControlledBy(&deploy, owner) {
			return &deploy, nil
		}
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func (dm DeploymentManager) Equal(a, b *appsv1.Deployment) bool {
	return reflect.DeepEqual(a.Spec, b.Spec) &&
		reflect.DeepEqual(a.Annotations, b.Annotations) &&
		reflect.DeepEqual(a.Labels, b.Labels)
}

func (dm DeploymentManager) Construct(provider *pipelinesv1.Provider) (*appsv1.Deployment, error) {
	prefixedProviderName := fmt.Sprintf("provider-%s", provider.Name)

	matchLabels := map[string]string{AppLabel: prefixedProviderName}
	deploymentLabels := MapConcat(dm.config.DefaultProviderValues.Labels, matchLabels)
	replicas := int32(dm.config.DefaultProviderValues.Replicas)

	podTemplate := dm.config.DefaultProviderValues.PodTemplateSpec
	populatedPodTemplate, err := populateServiceContainer(dm.config.DefaultProviderValues.ServiceContainerName, *podTemplate.DeepCopy(), provider)
	if err != nil {
		return nil, err
	}
	podTemplate = *populatedPodTemplate
	podTemplate.Spec.ServiceAccountName = provider.Spec.ServiceAccount
	podTemplate.ObjectMeta.Labels = MapConcat(podTemplate.ObjectMeta.Labels, matchLabels)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", prefixedProviderName),
			Namespace:    provider.Namespace,
			Labels:       deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Replicas: &replicas,
			Template: podTemplate,
		},
	}

	return deployment, nil
}

func populateServiceContainer(serviceContainerName string, podTemplate corev1.PodTemplateSpec, provider *pipelinesv1.Provider) (*corev1.PodTemplateSpec, error) {
	if !Exists(podTemplate.Spec.Containers, func(c corev1.Container) bool {
		return c.Name == serviceContainerName
	}) {

		return nil, fmt.Errorf("unable to populate service container: container with name %s not found on deployment", serviceContainerName)
	}

	paramsAsEnvVars := make([]corev1.EnvVar, len(provider.Spec.Parameters))
	for name, value := range provider.Spec.Parameters {
		paramsAsEnvVars = append(paramsAsEnvVars, corev1.EnvVar{Name: fmt.Sprintf("PARAMETERS_%s", strings.ToUpper(name)), Value: jsonToString(value)})
	}

	podTemplate.Spec.Containers = Map(podTemplate.Spec.Containers, func(c corev1.Container) corev1.Container {
		if c.Name == serviceContainerName {
			c.Image = provider.Spec.ServiceImage
			c.Env = append(c.Env,
				corev1.EnvVar{
					Name:  "PROVIDERNAME",
					Value: provider.Name,
				},
			)
			c.Env = append(c.Env, paramsAsEnvVars...)
		}
		return c
	})

	return &podTemplate, nil
}

func jsonToString(jsonValue *apiextensionsv1.JSON) string {
	var s string
	if jsonValue != nil {
		// Attempts to unmarshal input into a string to remove extra quotes and escape chars if the input is, in fact, a string
		if err := json.Unmarshal(jsonValue.Raw, &s); err != nil {
			s = string(jsonValue.Raw)
		}
	}
	return s
}
