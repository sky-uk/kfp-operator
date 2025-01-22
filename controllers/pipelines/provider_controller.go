package pipelines

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
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
	OwnerNameLabel         = "owner-name"
	AppLabel               = "app"
	ResourceHashAnnotation = "resource-hash"
)

type ProviderReconciler struct {
	StateHandler[*pipelinesv1.Provider]
	ResourceReconciler[*pipelinesv1.Provider]
	Scheme *runtime.Scheme
}

func NewProviderReconciler(ec K8sExecutionContext, config config.KfpControllerConfigSpec) *ProviderReconciler {
	return &ProviderReconciler{
		StateHandler: StateHandler[*pipelinesv1.Provider]{},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Provider]{
			EC:     ec,
			Config: config,
		},
		Scheme: ec.Scheme,
	}
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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

	desiredDeployment, err := r.constructDeployment(provider, req.Namespace, *r.Config.DeepCopy())
	if err != nil {
		logger.Error(err, "unable to construct provider service deployment")
		return ctrl.Result{}, err
	}

	existingDeployment, err := r.getDeployment(ctx, req.Namespace, provider.Name, *provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing deployment")
		return ctrl.Result{}, err
	}

	logger.Info("desired provider deployment", "deployment", desiredDeployment)

	if existingDeployment != nil {
		logger.Info("found existing provider deployment", "deployment", existingDeployment)

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

func (r *ProviderReconciler) constructDeployment(provider *pipelinesv1.Provider, namespace string, config config.KfpControllerConfigSpec) (*appsv1.Deployment, error) {
	matchLabels := map[string]string{AppLabel: fmt.Sprintf("provider-%s", provider.Name)}
	ownerLabels := map[string]string{OwnerNameLabel: fmt.Sprintf("provider-%s", provider.Name)}
	deploymentLabels := pipelines.MapConcat(pipelines.MapConcat(config.DefaultProviderValues.Labels, matchLabels), ownerLabels)
	replicas := int32(config.DefaultProviderValues.Replicas)

	podTemplate := config.DefaultProviderValues.PodTemplateSpec
	serviceContainerName := config.DefaultProviderValues.ServiceContainerName
	podTemplate = populateServiceContainer(serviceContainerName, *podTemplate.DeepCopy(), provider)
	podTemplate.Spec.ServiceAccountName = provider.Spec.ServiceAccount
	podTemplate.ObjectMeta.Labels = pipelines.MapConcat(podTemplate.ObjectMeta.Labels, matchLabels)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("provider-%s-", provider.Name),
			Namespace:    namespace,
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

	if err := ctrl.SetControllerReference(provider, deployment, r.Scheme); err != nil {
		return nil, err
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
	}
	return podTemplate
}

func setResourceHashAnnotation(deployment *appsv1.Deployment) error {
	hasher := pipelines.NewObjectHasher()
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

func (r *ProviderReconciler) getDeployment(ctx context.Context, namespace string, providerName string, provider pipelinesv1.Provider) (*appsv1.Deployment, error) {
	dl := &appsv1.DeploymentList{}
	err := r.EC.Client.NonCached.List(ctx, dl, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector(map[string]string{OwnerNameLabel: fmt.Sprintf("provider-%s", providerName)}),
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

func labelSelector(labelMap map[string]string) labels.Selector {
	return labels.SelectorFromSet(labelMap)
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
