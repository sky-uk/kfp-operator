package pipelines

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]
	DependingOnRunConfigurationReconciler[*pipelinesv1.RunConfiguration]
	EC     K8sExecutionContext
	Scheme *runtime.Scheme
	Config config.KfpControllerConfigSpec
}

func NewRunConfigurationReconciler(
	ec K8sExecutionContext,
	scheme *runtime.Scheme,
	config config.KfpControllerConfigSpec,
) *RunConfigurationReconciler {
	return &RunConfigurationReconciler{
		DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]{
			EC: ec,
		},
		DependingOnRunConfigurationReconciler[*pipelinesv1.RunConfiguration]{
			EC: ec,
		},
		ec,
		scheme,
		config,
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

	if runConfiguration.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	provider := runConfiguration.Spec.Run.Provider
	if runConfiguration.Status.Provider != "" && provider != runConfiguration.Status.Provider {
		//TODO: refactor to use Commands and introduce a StateHandler
		runConfiguration.Status.SynchronizationState = apis.Failed

		message := fmt.Sprintf(
			`%s: %s`,
			string(runConfiguration.Status.SynchronizationState),
			StateHandlerConstants.ProviderChangedError,
		)
		r.EC.Recorder.Event(runConfiguration, EventTypes.Warning, EventReasons.SyncFailed, message)

		return ctrl.Result{}, r.EC.Client.Status().Update(ctx, runConfiguration)
	}

	if hasChanged, err := r.handleObservedPipelineVersion(
		ctx,
		runConfiguration.Spec.Run.Pipeline,
		runConfiguration,
	); hasChanged || err != nil {
		return ctrl.Result{}, err
	}

	if hasChanged, err := r.handleDependentRuns(ctx, runConfiguration); hasChanged || err != nil {
		return ctrl.Result{}, err
	}

	var newStatus pipelinesv1.RunConfigurationStatus
	state := apis.Succeeded
	message := ""

	if resolvedParameters, err := runConfiguration.Spec.Run.ResolveRuntimeParameters(
		runConfiguration.Status.Dependencies,
	); err == nil {
		if hasChanged, err := r.syncWithRuns(ctx, runConfiguration); hasChanged || err != nil {
			return ctrl.Result{}, err
		}

		state, message, err = r.syncStatus(ctx, provider, runConfiguration, resolvedParameters)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	newStatus = runConfiguration.Status
	newStatus.ObservedGeneration = runConfiguration.GetGeneration()
	newStatus.Provider = provider

	newStatus.SetSynchronizationState(state, message)

	if !reflect.DeepEqual(newStatus, runConfiguration.Status) {
		runConfiguration.Status = newStatus
		if err := r.EC.Client.Status().Update(ctx, runConfiguration); err != nil {
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunConfigurationReconciler) triggerUntriggeredRuns(
	ctx context.Context,
	runConfiguration *pipelinesv1.RunConfiguration,
) error {
	runs, err := findOwnedRuns(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return err
	}

	desiredRun, err := r.constructRunForRunConfiguration(runConfiguration)
	if err != nil {
		return err
	}

	runExists := pipelines.Exists(runs, func(run pipelinesv1.Run) bool {
		return string(run.ComputeHash()) == string(desiredRun.ComputeHash())
	})

	if runExists {
		return nil
	}

	return r.EC.Client.Create(ctx, desiredRun)
}

func (r *RunConfigurationReconciler) updateRcTriggers(
	runConfiguration pipelinesv1.RunConfiguration,
) pipelinesv1.RunConfigurationStatus {
	newStatus := runConfiguration.Status

	if pipelines.Contains(runConfiguration.Spec.Triggers.OnChange, pipelinesv1.OnChangeTypes.Pipeline) {
		newStatus.TriggeredPipelineVersion = runConfiguration.Status.ObservedPipelineVersion
	}

	if pipelines.Contains(runConfiguration.Spec.Triggers.OnChange, pipelinesv1.OnChangeTypes.RunSpec) {
		newStatus.Triggers.RunSpec.Version = runConfiguration.Spec.Run.ComputeVersion()
	}

	newStatus.Triggers.RunConfigurations = pipelines.ToMap(
		runConfiguration.Spec.Triggers.RunConfigurations,
		func(rcName string) (string, pipelinesv1.TriggeredRunReference) {
			triggeredRun := pipelinesv1.TriggeredRunReference{
				ProviderId: runConfiguration.Status.Dependencies.RunConfigurations[rcName].ProviderId,
			}
			return rcName, triggeredRun
		},
	)

	return newStatus
}

func (r *RunConfigurationReconciler) syncWithRuns(
	ctx context.Context,
	runConfiguration *pipelinesv1.RunConfiguration,
) (bool, error) {
	oldStatus := runConfiguration.Status
	runConfiguration.Status = r.updateRcTriggers(*runConfiguration)

	if runConfiguration.Status.Triggers.Equals(oldStatus.Triggers) &&
		runConfiguration.Status.TriggeredPipelineVersion == oldStatus.TriggeredPipelineVersion {
		return false, nil
	}

	if err := r.triggerUntriggeredRuns(ctx, runConfiguration); err != nil {
		return false, err
	}

	return true, r.EC.Client.Status().Update(ctx, runConfiguration)
}

func (r *RunConfigurationReconciler) syncStatus(
	ctx context.Context,
	provider string,
	runConfiguration *pipelinesv1.RunConfiguration,
	resolvedParameters []apis.NamedValue,
) (state apis.SynchronizationState, message string, err error) {
	desiredSchedules, err := r.constructRunSchedulesForTriggers(runConfiguration, resolvedParameters)
	if err != nil {
		return
	}

	dependentSchedules, err := findOwnedRunSchedules(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return
	}

	missingSchedules := pipelines.SliceDiff(desiredSchedules, dependentSchedules, compareRunSchedules)
	excessSchedules := pipelines.SliceDiff(dependentSchedules, desiredSchedules, compareRunSchedules)
	excessSchedulesNotMarkedForDeletion := pipelines.Filter(
		excessSchedules, func(schedule pipelinesv1.RunSchedule) bool {
			return schedule.DeletionTimestamp == nil
		},
	)

	isSynced := len(missingSchedules) == 0 && len(excessSchedulesNotMarkedForDeletion) == 0

	if !isSynced {
		for _, desiredSchedule := range missingSchedules {
			if err = r.EC.Client.Create(ctx, &desiredSchedule); err != nil {
				return
			}
		}

		for _, excessSchedule := range excessSchedulesNotMarkedForDeletion {
			if err = r.EC.Client.Delete(ctx, &excessSchedule); err != nil {
				return
			}
		}

		return
	}

	state, message = aggregateState(dependentSchedules)

	return
}

func (r *RunConfigurationReconciler) reconciliationRequestsForPipeline(
	ctx context.Context,
	pipeline client.Object,
) []reconcile.Request {
	referencingRunConfigurations := &pipelinesv1.RunConfigurationList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(pipelineRefField, pipeline.GetName()),
		Namespace:     pipeline.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(ctx, referencingRunConfigurations, listOps)
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

func (r *RunConfigurationReconciler) reconciliationRequestsForRunConfiguration(
	ctx context.Context,
	runConfiguration client.Object,
) []reconcile.Request {
	referencingRunConfigurations := &pipelinesv1.RunConfigurationList{}
	rcRefListOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rcRefField, runConfiguration.GetName()),
		Namespace:     runConfiguration.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(ctx, referencingRunConfigurations, rcRefListOps)
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

	controllerBuilder, err := r.DependingOnPipelineReconciler.setupWithManager(
		mgr,
		controllerBuilder,
		runConfiguration,
		r.reconciliationRequestsForPipeline,
	)
	if err != nil {
		return err
	}
	controllerBuilder, err = r.DependingOnRunConfigurationReconciler.setupWithManager(
		mgr,
		controllerBuilder,
		runConfiguration,
		r.reconciliationRequestsForRunConfiguration,
	)
	if err != nil {
		return err
	}

	return controllerBuilder.Owns(&pipelinesv1.RunSchedule{}).Complete(r)
}

func findOwnedRunSchedules(
	ctx context.Context,
	cli client.Reader,
	runConfiguration *pipelinesv1.RunConfiguration,
) ([]pipelinesv1.RunSchedule, error) {
	ownedRunSchedulesList := &pipelinesv1.RunScheduleList{}
	if err := cli.List(
		ctx,
		ownedRunSchedulesList,
		client.InNamespace(runConfiguration.Namespace),
	); err != nil {
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

func findOwnedRuns(
	ctx context.Context,
	cli client.Reader,
	runConfiguration *pipelinesv1.RunConfiguration,
) ([]pipelinesv1.Run, error) {
	ownedRunsList := &pipelinesv1.RunList{}
	if err := cli.List(ctx, ownedRunsList, client.InNamespace(runConfiguration.Namespace)); err != nil {
		return nil, err
	}

	var ownedRuns []pipelinesv1.Run
	for _, run := range ownedRunsList.Items {
		if metav1.IsControlledBy(&run, runConfiguration) {
			ownedRuns = append(ownedRuns, run)
		}
	}

	return ownedRuns, nil
}

func (r *RunConfigurationReconciler) constructRunForRunConfiguration(
	runConfiguration *pipelinesv1.RunConfiguration,
) (*pipelinesv1.Run, error) {
	spec := runConfiguration.Spec.Run
	spec.Pipeline.Version = runConfiguration.Status.ObservedPipelineVersion

	run := pipelinesv1.Run{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: runConfiguration.Name + "-",
			Namespace:    runConfiguration.Namespace,
			Labels: map[string]string{
				workflowfactory.RunConfigurationConstants.RunConfigurationNameLabelKey: runConfiguration.GetName(),
			},
		},
		Spec: spec,
	}

	if err := controllerutil.SetControllerReference(runConfiguration, &run, r.Scheme); err != nil {
		return nil, err
	}

	return &run, nil
}

func (r *RunConfigurationReconciler) constructRunSchedulesForTriggers(
	runConfiguration *pipelinesv1.RunConfiguration,
	resolvedParameters []apis.NamedValue,
) ([]pipelinesv1.RunSchedule, error) {
	var schedules []pipelinesv1.RunSchedule

	for _, schedule := range runConfiguration.Spec.Triggers.Schedules {
		runSchedule := pipelinesv1.RunSchedule{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: runConfiguration.Name + "-",
				Namespace:    runConfiguration.Namespace,
			},
			Spec: pipelinesv1.RunScheduleSpec{
				Provider: runConfiguration.Spec.Run.Provider,
				Pipeline: pipelinesv1.PipelineIdentifier{
					Name:    runConfiguration.Spec.Run.Pipeline.Name,
					Version: runConfiguration.Status.ObservedPipelineVersion,
				},
				RuntimeParameters: resolvedParameters,
				ExperimentName:    runConfiguration.Spec.Run.ExperimentName,
				Artifacts:         runConfiguration.Spec.Run.Artifacts,
				Schedule:          schedule,
			},
		}
		if err := controllerutil.SetControllerReference(
			runConfiguration,
			&runSchedule,
			r.Scheme,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, runSchedule)
	}

	return schedules, nil
}

func aggregateState(dependencies []pipelinesv1.RunSchedule) (aggState apis.SynchronizationState, message string) {
	aggState = apis.Succeeded

	for _, dependency := range dependencies {
		if dependency.Status.SynchronizationState == apis.Failed {
			aggState = apis.Failed
			message = dependency.Status.Conditions.SynchronizationSucceeded().Message
		} else if dependency.Status.SynchronizationState != apis.Succeeded {
			aggState = apis.Updating
			message = ""
			return
		}
	}

	return aggState, message
}

func compareRunSchedules(a, b pipelinesv1.RunSchedule) bool {
	return string(a.ComputeHash()) == string(b.ComputeHash())
}
