package pipelines

import (
	"context"
	"time"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ProviderReconciler reconciles a Provider object
type ProviderReconciler struct {
	StateHandler[*pipelinesv1.Provider]
	ResourceReconciler[*pipelinesv1.Provider]
}

func NewProviderReconciler(ec K8sExecutionContext) *ProviderReconciler {
	return &ProviderReconciler{
		StateHandler: StateHandler[*pipelinesv1.Provider]{},
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Provider]{
			EC: ec,
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

	logger.V(3).Info("found provider", "resource", provider)

	// commands := r.StateHandler.StateTransition(ctx, provider, provider)

	// for i := range commands {
	// 	if err := commands[i].execute(ctx, r.EC, provider); err != nil {
	// 		logger.Error(err, "error executing command", LogKeys.Command, commands[i])
	// 		return ctrl.Result{}, err
	// 	}
	// }

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelinesv1.Provider{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(provider)

	controllerBuilder = r.ResourceReconciler.setupWithManager(controllerBuilder, provider)

	return controllerBuilder.Complete(r)
}
