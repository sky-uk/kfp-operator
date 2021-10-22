package pipelines

import (
	"context"
	"fmt"
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
	Scheme       *runtime.Scheme
	StateHandler RunConfigurationStateHandler
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/finalizers,verbs=update

func (r *RunConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var runConfiguration = &pipelinesv1.RunConfiguration{}
	if err := r.Get(ctx, req.NamespacedName, runConfiguration); err != nil {
		logger.Error(err, "unable to fetch run configuration")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if runConfiguration.ObjectMeta.DeletionTimestamp.IsZero() {
		r.AddFinalizer(ctx, runConfiguration)
	}

	commands := r.StateHandler.StateTransition(ctx, runConfiguration)

	for i := range commands {
		if err := commands[i].execute(r, ctx, runConfiguration); err != nil {
			logger.Error(err, fmt.Sprintf("Error executing command: %+v", commands[i]))
			return ctrl.Result{}, err
		}
	}

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

func (r *RunConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1.RunConfiguration{}).
		Owns(&argo.Workflow{}).
		Complete(r)
}
