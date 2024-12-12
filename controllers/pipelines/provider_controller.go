package pipelines

import (
	"context"
	"fmt"
	"time"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ProviderReconciler struct {
	StateHandler[*pipelinesv1.Provider]
	ResourceReconciler[*pipelinesv1.Provider]
}

func NewProviderReconciler(ec K8sExecutionContext, config config.KfpControllerConfigSpec) *ProviderReconciler {
	return &ProviderReconciler{
		StateHandler: StateHandler[*pipelinesv1.Provider]{},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Provider]{
			EC:     ec,
			Config: config,
		},
	}
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var provider = &pipelinesv1.Provider{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, provider); err != nil {
		logger.Error(err, "unable to fetch provider")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	providerVersion := provider.ComputeVersion()
	logger.V(2).Info("found provider", "resource", provider, "version", providerVersion)

	providerServiceName := fmt.Sprintf("provider-%s-service-%s", req.Name, providerVersion)
	providerServiceDeployment, err := r.fetchProviderServiceDeployment(ctx, providerServiceName, req.Namespace)
	if err != nil {
		logger.Error(err, "unable to fetch provider service deployment")
		return ctrl.Result{}, err
	}

	if providerServiceDeployment == nil {
		providerServiceDeployment = constructProviderDeployment(providerServiceName, req.Namespace, provider, r.Config)
		if err := r.EC.Client.Create(ctx, providerServiceDeployment); err != nil {
			logger.Error(err, "unable to create provider service deployment")
			return ctrl.Result{}, err
		}
		logger.Info("created provider service deployment", "resource", providerServiceDeployment)
	} else {
		logger.V(2).Info("found provider service deployment", "resource", providerServiceDeployment)
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelinesv1.Provider{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(provider)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, provider)

	return controllerBuilder.Complete(r)
}

func (r *ProviderReconciler) fetchProviderServiceDeployment(ctx context.Context, name, namespace string) (*appsv1.Deployment, error) {
	var deployment = &appsv1.Deployment{}
	err := r.EC.Client.NonCached.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, deployment)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return deployment, nil
}

func constructProviderDeployment(name, namespace string, provider *pipelinesv1.Provider, config config.KfpControllerConfigSpec) *appsv1.Deployment {
	replicas := int32(config.DefaultProviderValues.Replicas)

	podTemplate := config.DefaultProviderValues.PodTemplateSpec

	const targetContainer = "provider-service" //TODO: make configurable from config
	for _, container := range podTemplate.Spec.Containers {
		if container.Name == targetContainer {
			container.Image = provider.Spec.Image
			container.Env = append(container.Env, v1.EnvVar{
				Name:  "PROVIDERNAME",
				Value: provider.Name,
			})
		}
	}

	const providerLabelKey = "provider"
	podTemplate.ObjectMeta.Labels[providerLabelKey] = provider.Name
	podTemplate.Spec.ServiceAccountName = provider.Spec.ServiceAccount

	labels := config.DefaultProviderValues.Labels
	labels[providerLabelKey] = provider.Name

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("provider-%s", name),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: podTemplate,
		},
	}
}
