package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RunConfiguration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *RunConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *RunConfigurationReconciler) AddFinalizer(ctx context.Context, runconfiguration *pipelinesv1.RunConfiguration) error {
	if !containsString(runconfiguration.ObjectMeta.Finalizers, finalizerName) {
		runconfiguration.ObjectMeta.Finalizers = append(runconfiguration.ObjectMeta.Finalizers, finalizerName)
		return r.Update(ctx, runconfiguration)
	}

	return nil
}

func (r *RunConfigurationReconciler) RemoveFinalizer(ctx context.Context, runconfiguration pipelinesv1.RunConfiguration) error {
	if containsString(runconfiguration.ObjectMeta.Finalizers, finalizerName) {
		runconfiguration.ObjectMeta.Finalizers = removeString(runconfiguration.ObjectMeta.Finalizers, finalizerName)
		return r.Update(ctx, &runconfiguration)
	}

	return nil
}

func (r *RunConfigurationReconciler) CreateChildWorkflow(ctx context.Context, runconfiguration *pipelinesv1.RunConfiguration, workflow argo.Workflow) error {
	if err := ctrl.SetControllerReference(runconfiguration, &workflow, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, &workflow); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RunConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.RunConfiguration{}).
		Complete(r)
}
