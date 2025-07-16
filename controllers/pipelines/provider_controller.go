package pipelines

import (
	"context"
	"errors"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/controllerconfigutil"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	AppLabel = "app"
)

type ProviderReconciler struct {
	ProviderLoader    ProviderLoader
	DeploymentManager DeploymentResourceManager
	ServiceManager    ServiceResourceManager
	StatusManager     ProviderStatusManager
}

func NewProviderReconciler(ec K8sExecutionContext, config config.KfpControllerConfigSpec) *ProviderReconciler {
	return &ProviderReconciler{
		ProviderLoader: ResourceReconciler[*pipelineshub.Provider]{
			EC:     ec,
			Config: config,
		},
		DeploymentManager: DeploymentManager{
			scheme: ec.Scheme,
			client: &ec.Client,
			config: &config,
		},
		ServiceManager: ServiceManager{
			scheme: ec.Scheme,
			client: &ec.Client,
			config: &config,
		},
		StatusManager: StatusManager{
			client: &ec.Client,
		},
	}
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelineshub.Provider{}
	return ctrl.NewControllerManagedBy(mgr).
		For(provider, builder.WithPredicates(
			predicate.GenerationChangedPredicate{},
		)).
		WithOptions(controller.Options{
			RateLimiter: controllerconfigutil.RateLimiter,
		}).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(
			predicate.GenerationChangedPredicate{},
			predicate.ResourceVersionChangedPredicate{},
		)).
		Owns(&v1.Service{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started", "request", req)

	provider, err := r.ProviderLoader.LoadProvider(ctx, common.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	})
	if err != nil {
		logger.Error(err, "unable to get provider", "provider", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	providerOutOfSync := provider.Status.ObservedGeneration != provider.Generation

	existingDeployment, err := r.DeploymentManager.Get(ctx, &provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing deployment", "provider", provider.GetNamespacedName())
		return ctrl.Result{}, nil
	}

	desiredDeployment, err := r.DeploymentManager.Construct(&provider)
	if err != nil {
		logger.Error(err, "unable to construct provider deployment", "provider", provider.GetNamespacedName())
		return ctrl.Result{}, nil
	}

	if providerOutOfSync || existingDeployment == nil || !r.DeploymentManager.Equal(existingDeployment, desiredDeployment) {
		if existingDeployment == nil {
			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Creating, "")

			if err := r.DeploymentManager.Create(ctx, desiredDeployment, &provider); err != nil {
				if statusError := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource deployment"); statusError != nil {
					err = errors.Join(err, statusError)
				}
				return ctrl.Result{}, err
			}

		} else {
			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Updating, "")

			if err := r.DeploymentManager.Update(ctx, existingDeployment, desiredDeployment, &provider); err != nil {
				if statusError := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource deployment"); statusError != nil {
					err = errors.Join(err, statusError)
				}
				return ctrl.Result{}, err
			}
		}
	}

	existingSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing service", "provider", provider.GetNamespacedName())
		return ctrl.Result{}, nil
	}

	desiredSvc := r.ServiceManager.Construct(&provider)

	if providerOutOfSync || existingSvc == nil || !r.ServiceManager.Equal(existingSvc, desiredSvc) {
		if existingSvc == nil {

			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Creating, "")

			if err := r.ServiceManager.Create(ctx, desiredSvc, &provider); err != nil {
				if statusError := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource service"); statusError != nil {
					err = errors.Join(err, statusError)
				}
				return ctrl.Result{}, err
			}

		} else {

			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Updating, "")

			// delete, to allow a new one to be created avoiding issues with immutability
			if err := r.ServiceManager.Delete(ctx, existingSvc); err != nil {
				if statusError := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource service"); statusError != nil {
					err = errors.Join(err, statusError)
				}
				return ctrl.Result{}, err
			}
		}
	}

	if err := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Succeeded, ""); err != nil {
		return ctrl.Result{}, err
	}

	duration := time.Since(startTime)
	logger.Info("Provider reconciliation ended", "provider", req.String(), logkeys.Duration, duration)

	return ctrl.Result{}, nil
}
