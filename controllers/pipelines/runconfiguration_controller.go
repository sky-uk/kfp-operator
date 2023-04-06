package pipelines

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runconfigurations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipeline,verbs=get;list;watch

func (r *RunConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()
	logger.V(2).Info("reconciliation started")

	var runConfiguration = &pipelinesv1.RunConfiguration{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, runConfiguration); err != nil {
		logger.Error(err, "unable to fetch run configuration")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run configuration", "resource", runConfiguration)

	desiredProvider := r.desiredProvider(runConfiguration)

	if hasChanged, err := r.handleObservedPipelineVersion(ctx, runConfiguration.Spec.Pipeline, runConfiguration); hasChanged || err != nil {
		return ctrl.Result{}, err
	}

	commands := r.StateHandler.StateTransition(ctx, desiredProvider, runConfiguration)

	for i := range commands {
		if err := commands[i].execute(ctx, r.EC, runConfiguration); err != nil {
			logger.Error(err, "error executing command", LogKeys.Command, commands[i])
			return ctrl.Result{}, err
		}
	}

	if err := r.syncRunSchedule(ctx, runConfiguration); err != nil {
		return ctrl.Result{}, err
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunConfigurationReconciler) syncRunSchedule(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) error {
	expectedRunSchedule := runScheduleForRunConfiguration(runConfiguration)
	ownedRunSchedule := &pipelinesv1.RunSchedule{}
	err := r.EC.Client.NonCached.Get(ctx, runConfiguration.GetNamespacedName(), ownedRunSchedule)

	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		if err = controllerutil.SetControllerReference(runConfiguration, expectedRunSchedule, r.EC.Scheme); err != nil {
			return err
		}

		if err = r.EC.Client.Create(ctx, expectedRunSchedule); err != nil {
			return err
		}
	} else {
		ownedRunSchedule.Spec = expectedRunSchedule.Spec
		if err = r.EC.Client.Update(ctx, ownedRunSchedule); err != nil {
			return err
		}
	}

	if err := r.EC.Client.NonCached.Get(ctx, runConfiguration.GetNamespacedName(), ownedRunSchedule); err != nil {
		return err
	}

	ownedRunSchedule.Status = expectedRunSchedule.Status

	return r.EC.Client.Status().Update(ctx, ownedRunSchedule)
}

func runScheduleForRunConfiguration(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunSchedule {
	rs := &pipelinesv1.RunSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runConfiguration.Name,
			Namespace: runConfiguration.Namespace,
		},
		Spec: pipelinesv1.RunScheduleSpec{
			Schedule:       runConfiguration.Spec.Schedule,
			ExperimentName: runConfiguration.Spec.ExperimentName,
			Pipeline: pipelinesv1.PipelineIdentifier{
				Name:    runConfiguration.Spec.Pipeline.Name,
				Version: runConfiguration.Status.ObservedPipelineVersion,
			},
			RuntimeParameters: runConfiguration.Spec.RuntimeParameters,
		},
		Status: pipelinesv1.Status{
			ProviderId:           runConfiguration.Status.ProviderId,
			SynchronizationState: runConfiguration.Status.SynchronizationState,
		},
	}
	rs.Status.Version = rs.ComputeVersion()
	return rs
}

func (r *RunConfigurationReconciler) reconciliationRequestsForPipeline(pipeline client.Object) []reconcile.Request {
	referencingRunConfigurations := &pipelinesv1.RunConfigurationList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(pipelineRefField, pipeline.GetName()),
		Namespace:     pipeline.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(context.TODO(), referencingRunConfigurations, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(referencingRunConfigurations.Items))
	for i, item := range referencingRunConfigurations.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *RunConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder, err := r.setupWithManager(mgr, &pipelinesv1.RunConfiguration{}, r.reconciliationRequestsForPipeline)
	if err != nil {
		return err
	}

	return controllerBuilder.Complete(r)
}
