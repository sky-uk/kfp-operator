package pipelines

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	logger.V(2).Info("reconciliation started")

	var provider = &pipelinesv1.Provider{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, provider); err != nil {
		logger.Error(err, "unable to fetch provider")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("found provider", "resource", provider)

	desiredProviderDeployment, err := r.constructProviderServiceDeployment(provider, req.Namespace, r.Config)
	if err != nil {
		logger.Error(err, "unable to construct provider service deployment")
		return ctrl.Result{}, err
	}
	logger.Info("created provider deployment", "deployment", desiredProviderDeployment)

	if err := r.EC.Client.Patch(ctx, desiredProviderDeployment, client.Apply, client.FieldOwner("provider-controller-provider")); err != nil {
		logger.Error(err, "unable to update provider service deployment")
		return ctrl.Result{}, err
	}

	duration := time.Since(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelinesv1.Provider{}
	return ctrl.NewControllerManagedBy(mgr).
		For(provider).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *ProviderReconciler) constructProviderServiceDeployment(provider *pipelinesv1.Provider, namespace string, config config.KfpControllerConfigSpec) (*appsv1.Deployment, error) {
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

	podTemplate.Spec.ServiceAccountName = provider.Spec.ServiceAccount
	labels := config.DefaultProviderValues.Labels

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("provider-%s-", provider.Name),
			Namespace:    namespace,
			Labels:       labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: podTemplate,
		},
	}
	err := ctrl.SetControllerReference(provider, deployment, r.Scheme)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
