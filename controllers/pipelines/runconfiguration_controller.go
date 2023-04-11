package pipelines

import (
	"context"
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
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=pipelines,verbs=get;list;watch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=runschedules/status,verbs=get;update;patch

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
	ownedRunSchedules, err := findOwnedRunSchedules(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return err
	}

	var runSchedule *pipelinesv1.RunSchedule = nil

	for _, ownedRunSchedule := range ownedRunSchedules {
		if runSchedule == nil && match(ownedRunSchedule, *expectedRunSchedule) {
			runSchedule = &ownedRunSchedule
		} else {
			if err = r.EC.Client.Delete(ctx, &ownedRunSchedule); err != nil {
				return err
			}
		}
	}

	if runSchedule == nil {
		if err = controllerutil.SetControllerReference(runConfiguration, expectedRunSchedule, r.EC.Scheme); err != nil {
			return err
		}

		runSchedule = expectedRunSchedule

		if err = r.EC.Client.Create(ctx, runSchedule); err != nil {
			return err
		}
	}

	runSchedule.Status = expectedRunSchedule.Status

	return r.EC.Client.Status().Update(ctx, runSchedule)
}

func findOwnedRunSchedules(ctx context.Context, cli client.Reader, runConfiguration *pipelinesv1.RunConfiguration) ([]pipelinesv1.RunSchedule, error) {
	ownedRunSchedulesList := &pipelinesv1.RunScheduleList{}
	if err := cli.List(ctx, ownedRunSchedulesList, client.InNamespace(runConfiguration.Namespace)); err != nil {
		return nil, err
	}

	var ownedSchedules []pipelinesv1.RunSchedule
	for _, schedule := range ownedRunSchedulesList.Items {
		if metav1.IsControlledBy(&schedule, runConfiguration) {
			ownedSchedules = append(ownedSchedules, schedule)
		}
	}

	return ownedSchedules, nil
}

func runScheduleForRunConfiguration(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunSchedule {
	rs := &pipelinesv1.RunSchedule{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: runConfiguration.Name + "-",
			Namespace:    runConfiguration.Namespace,
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

func match(a, b pipelinesv1.RunSchedule) bool {
	return string(a.ComputeHash()) == string(b.ComputeHash())
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

	return controllerBuilder.Owns(&pipelinesv1.RunSchedule{}).Complete(r)
}
