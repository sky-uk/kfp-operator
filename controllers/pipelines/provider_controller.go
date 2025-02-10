package pipelines

import (
	"context"
	"errors"
	"fmt"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/predicates"
	"time"

	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
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
		ProviderLoader: ResourceReconciler[*pipelinesv1.Provider]{
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

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelinesv1.Provider{}
	return ctrl.NewControllerManagedBy(mgr).
		For(provider, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(
			predicate.GenerationChangedPredicate{},
			predicates.DeploymentChangedPredicate{},
		)).
		Owns(&v1.Service{}, builder.WithPredicates(
			predicates.ServiceChangedPredicate{},
		)).
		Complete(r)
}

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started", "request", req)

	provider, err := r.ProviderLoader.LoadProvider(ctx, req.Namespace, req.Name)
	if err != nil {
		logger.Error(err, "unable to get provider")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	providerOutOfSync := provider.Status.ObservedGeneration != provider.Generation

	existingDeployment, err := r.DeploymentManager.Get(ctx, &provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing deployment")
		return ctrl.Result{}, nil
	}

	desiredDeployment, err := r.DeploymentManager.Construct(&provider)
	if err != nil {
		logger.Error(err, "unable to construct provider deployment")
		return ctrl.Result{}, nil
	}

	if providerOutOfSync || existingDeployment == nil || !r.DeploymentManager.Equal(existingDeployment, desiredDeployment) {
		if existingDeployment == nil {

			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Creating, "Failed to reconcile subresource deployment")

			if err := r.DeploymentManager.Create(ctx, desiredDeployment, &provider); err != nil {
				if err2 := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource deployment"); err2 != nil {
					err = errors.Join(err, err2)
				}
				return ctrl.Result{}, err
			}
		} else {

			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Updating, "Failed to reconcile subresource deployment")

			if err := r.DeploymentManager.Update(ctx, existingDeployment, desiredDeployment, &provider); err != nil {
				if err2 := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource deployment"); err2 != nil {
					err = errors.Join(err, err2)
				}
				return ctrl.Result{}, err
			}
		}
	}

	existingSvc, err := r.ServiceManager.Get(ctx, &provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing service")
		return ctrl.Result{}, nil
	}

	desiredSvc, err := r.ServiceManager.Construct(&provider)
	if err != nil {
		logger.Error(err, "unable to construct provider service")
		return ctrl.Result{}, nil
	}

	if providerOutOfSync || existingSvc == nil || !r.ServiceManager.Equal(existingSvc, desiredSvc) {
		if existingSvc == nil {

			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Creating, "Failed to reconcile subresource deployment")

			if err := r.ServiceManager.Create(ctx, desiredSvc, &provider); err != nil {
				if err2 := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource service"); err2 != nil {
					err = errors.Join(err, err2)
				}
				return ctrl.Result{}, err
			}

		} else {

			r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Updating, "Failed to reconcile subresource deployment")

			// delete, to allow a new one to be created avoiding issues with immutability
			if err := r.ServiceManager.Delete(ctx, existingSvc); err != nil {
				if err2 := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Failed, "Failed to reconcile subresource service"); err2 != nil {
					err = errors.Join(err, err2)
				}
				return ctrl.Result{}, err
			}
		}
	}

	if err := r.StatusManager.UpdateProviderStatus(ctx, &provider, apis.Succeeded, ""); err != nil {
		return ctrl.Result{}, err
	}

	duration := time.Since(startTime)
	logger.Info(fmt.Sprintf("provider %s reconciliation successful", provider.Name), logkeys.Duration, duration)

	return ctrl.Result{}, nil
}
