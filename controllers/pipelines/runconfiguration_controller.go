package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]
	Scheme *runtime.Scheme
}

func NewRunConfigurationReconciler(ec K8sExecutionContext, scheme *runtime.Scheme) *RunConfigurationReconciler {
	return &RunConfigurationReconciler{
		DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]{
			EC: ec,
		},
		scheme,
	}
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

	if hasChanged, err := r.handleObservedPipelineVersion(ctx, runConfiguration.Spec.Run.Pipeline, runConfiguration); hasChanged || err != nil {
		return ctrl.Result{}, err
	}

	newStatus := runConfiguration.Status
	newStatus.ObservedGeneration = runConfiguration.GetGeneration()

	newSynchronizationState, err := r.syncWithRunSchedules(ctx, runConfiguration)
	if err != nil {
		return ctrl.Result{}, err
	}

	newStatus.SynchronizationState = newSynchronizationState

	if newStatus != runConfiguration.Status {
		runConfiguration.Status = newStatus
		if err := r.EC.Client.Status().Update(ctx, runConfiguration); err != nil {
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunConfigurationReconciler) syncWithRunSchedules(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (state apis.SynchronizationState, err error) {
	state = runConfiguration.Status.SynchronizationState

	desiredSchedules, err := constructRunSchedulesForTriggers(runConfiguration, r.Scheme)
	if err != nil {
		return
	}

	dependentSchedules, err := findOwnedRunSchedules(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return
	}

	missingSchedules := sliceDiff(desiredSchedules, dependentSchedules, compareRunSchedules)
	excessSchedules := sliceDiff(dependentSchedules, desiredSchedules, compareRunSchedules)
	isSynced := len(missingSchedules) == 0 && len(excessSchedules) == 0

	if !isSynced {
		for _, desiredSchedule := range missingSchedules {
			if err = r.EC.Client.Create(ctx, &desiredSchedule); err != nil {
				return
			}
		}

		for _, excessSchedule := range excessSchedules {
			if err = r.EC.Client.Delete(ctx, &excessSchedule); err != nil {
				return
			}
		}

		return
	}

	state = aggregateState(dependentSchedules)
	return
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
	runConfiguration := &pipelinesv1.RunConfiguration{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(runConfiguration)

	controllerBuilder, err := r.setupWithManager(mgr, controllerBuilder, runConfiguration, r.reconciliationRequestsForPipeline)
	if err != nil {
		return err
	}

	return controllerBuilder.Owns(&pipelinesv1.RunSchedule{}).Complete(r)
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

func constructRunSchedulesForTriggers(runConfiguration *pipelinesv1.RunConfiguration, scheme *runtime.Scheme) ([]pipelinesv1.RunSchedule, error) {
	var schedules []pipelinesv1.RunSchedule

	for _, trigger := range runConfiguration.Spec.Triggers {
		if trigger.Type == pipelinesv1.TriggerTypes.Schedule {
			schedule := pipelinesv1.RunSchedule{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: runConfiguration.Name + "-",
					Namespace:    runConfiguration.Namespace,
				},
				Spec: pipelinesv1.RunScheduleSpec{
					Pipeline: pipelinesv1.PipelineIdentifier{
						Name:    runConfiguration.Spec.Run.Pipeline.Name,
						Version: runConfiguration.Status.ObservedPipelineVersion,
					},
					RuntimeParameters: runConfiguration.Spec.Run.RuntimeParameters,
					Schedule:          trigger.CronExpression,
					ExperimentName:    runConfiguration.Spec.Run.ExperimentName,
				},
			}
			if err := controllerutil.SetControllerReference(runConfiguration, &schedule, scheme); err != nil {
				return nil, err
			}
			metav1.SetMetaDataAnnotation(&schedule.ObjectMeta, apis.ResourceAnnotations.Provider, runConfiguration.GetAnnotations()[apis.ResourceAnnotations.Provider])
			schedules = append(schedules, schedule)
		}
	}

	return schedules, nil
}

func aggregateState(dependencies []pipelinesv1.RunSchedule) apis.SynchronizationState {
	aggState := apis.Succeeded

	for _, dependency := range dependencies {
		if dependency.Status.SynchronizationState == apis.Failed {
			aggState = apis.Failed
		} else if dependency.Status.SynchronizationState != apis.Succeeded {
			return apis.Updating
		}
	}

	return aggState
}

func compareRunSchedules(a, b pipelinesv1.RunSchedule) bool {
	return string(a.ComputeHash()) == string(b.ComputeHash())
}
