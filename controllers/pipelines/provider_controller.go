package pipelines

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	. "github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	AppLabel               = "app"
	ResourceHashAnnotation = "resource-hash"
)

type ProviderReconciler struct {
	ResourceReconciler[*pipelinesv1.Provider]
	Scheme *runtime.Scheme
}

func NewProviderReconciler(ec K8sExecutionContext, config config.KfpControllerConfigSpec) *ProviderReconciler {
	return &ProviderReconciler{
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Provider]{
			EC:     ec,
			Config: config,
		},
		Scheme: ec.Scheme,
	}
}

// +kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.Info("reconciliation started", "request", req)

	var provider = &pipelinesv1.Provider{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, provider); err != nil {
		logger.Error(err, "unable to get provider")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("found provider", "resource", provider)

	desiredDeployment, err := constructDeployment(provider, *r.Config.DeepCopy())
	if err != nil {
		logger.Error(err, "unable to construct provider deployment")
		return ctrl.Result{}, err
	}
	if err := ctrl.SetControllerReference(provider, desiredDeployment, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference on deployment")
		return ctrl.Result{}, err
	}

	logger.Info("constructed desired provider deployment", "deployment", desiredDeployment)

	existingDeployment, err := r.getDeployment(ctx, *provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing deployment")
		return ctrl.Result{}, err
	}

	if existingDeployment != nil {
		logger.Info("found existing provider deployment", "deployment", existingDeployment)

		if err := setResourceHashAnnotation(existingDeployment); err != nil {
			logger.Error(err, "unable to set resource hash annotation on existing deployment")
			return ctrl.Result{}, err
		}
		if deploymentIsOutOfSync(existingDeployment, desiredDeployment) {
			logger.Info("resource hash mismatch, updating deployment")
			existingDeployment = syncDeployment(existingDeployment, desiredDeployment)
			if err = r.EC.Client.Update(ctx, existingDeployment); err != nil {
				logger.Error(err, "unable to update provider service deployment", "deployment", desiredDeployment)
				return ctrl.Result{}, err
			}
		}
	} else {
		if err = r.EC.Client.Create(ctx, desiredDeployment); err != nil {
			logger.Error(err, "unable to create provider service deployment")
			return ctrl.Result{}, err
		}

		logger.Info("created provider deployment", "deployment", desiredDeployment)
	}

	duration := time.Since(startTime)
	logger.Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelinesv1.Provider{}
	return ctrl.NewControllerManagedBy(mgr).
		For(provider).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func constructDeployment(provider *pipelinesv1.Provider, config config.KfpControllerConfigSpec) (*appsv1.Deployment, error) {
	matchLabels := map[string]string{AppLabel: fmt.Sprintf("provider-%s", provider.Name)}
	deploymentLabels := MapConcat(config.DefaultProviderValues.Labels, matchLabels)
	replicas := int32(config.DefaultProviderValues.Replicas)

	podTemplate := config.DefaultProviderValues.PodTemplateSpec
	serviceContainerName := config.DefaultProviderValues.ServiceContainerName
	podTemplate = populateServiceContainer(serviceContainerName, *podTemplate.DeepCopy(), provider)
	podTemplate.Spec.ServiceAccountName = provider.Spec.ServiceAccount
	podTemplate.ObjectMeta.Labels = MapConcat(podTemplate.ObjectMeta.Labels, matchLabels)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("provider-%s-", provider.Name),
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

	if err := setResourceHashAnnotation(deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func populateServiceContainer(serviceContainerName string, podTemplate v1.PodTemplateSpec, provider *pipelinesv1.Provider) v1.PodTemplateSpec {
	for i, container := range podTemplate.Spec.Containers {
		if container.Name == serviceContainerName {
			podTemplate.Spec.Containers[i].Image = provider.Spec.ServiceImage
			podTemplate.Spec.Containers[i].Env = append(podTemplate.Spec.Containers[i].Env, v1.EnvVar{
				Name:  "PROVIDERNAME",
				Value: provider.Name,
			})
		}
		break
	}
	return podTemplate
}

func setResourceHashAnnotation(deployment *appsv1.Deployment) error {
	hasher := NewObjectHasher()
	err := hasher.WriteObject(deployment)
	if err != nil {
		return err
	}

	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations[ResourceHashAnnotation] = fmt.Sprintf("%x", hasher.Sum())

	return nil
}

func (r *ProviderReconciler) getDeployment(ctx context.Context, provider pipelinesv1.Provider) (*appsv1.Deployment, error) {
	dl := &appsv1.DeploymentList{}
	err := r.EC.Client.NonCached.List(ctx, dl, &client.ListOptions{
		Namespace: provider.Namespace,
	})
	if err != nil {
		return nil, err
	}

	for _, deploy := range dl.Items {
		if metav1.IsControlledBy(&deploy, &provider) {
			return &deploy, nil
		}
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func deploymentIsOutOfSync(existingDeployment, desiredDeployment *appsv1.Deployment) bool {
	return existingDeployment.Annotations != nil && existingDeployment.Annotations[ResourceHashAnnotation] != desiredDeployment.Annotations[ResourceHashAnnotation]
}

func syncDeployment(existingDeployment, desiredDeployment *appsv1.Deployment) *appsv1.Deployment {
	syncedDeployment := existingDeployment.DeepCopy()

	syncedDeployment.Spec = desiredDeployment.Spec
	syncedDeployment.SetLabels(desiredDeployment.Labels)
	syncedDeployment.Annotations[ResourceHashAnnotation] = desiredDeployment.Annotations[ResourceHashAnnotation]

	return syncedDeployment
}
