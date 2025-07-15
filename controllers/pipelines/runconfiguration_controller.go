package pipelines

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/common/triggers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/controllerconfigutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"slices"
	"time"

	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	DependingOnPipelineReconciler[*pipelineshub.RunConfiguration]
	DependingOnRunConfigurationReconciler[*pipelineshub.RunConfiguration]
	EC     K8sExecutionContext
	Scheme *runtime.Scheme
	Config config.KfpControllerConfigSpec
}

func NewRunConfigurationReconciler(
	ec K8sExecutionContext,
	config config.KfpControllerConfigSpec,
) *RunConfigurationReconciler {
	return &RunConfigurationReconciler{
		DependingOnPipelineReconciler[*pipelineshub.RunConfiguration]{
			EC: ec,
		},
		DependingOnRunConfigurationReconciler[*pipelineshub.RunConfiguration]{
			EC: ec,
		},
		ec,
		ec.Scheme,
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

	var runConfiguration = &pipelineshub.RunConfiguration{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, runConfiguration); err != nil {
		logger.Error(err, "unable to fetch run configuration")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(3).Info("found run configuration", "resource", runConfiguration)

	if runConfiguration.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	provider := runConfiguration.Spec.Run.Provider
	if !runConfiguration.Status.Provider.Empty() && provider != runConfiguration.Status.Provider {
		//TODO: refactor to use Commands and introduce a StateHandler
		runConfiguration.Status.SetSynchronizationState(apis.Failed, StateHandlerConstants.ProviderChangedError)

		message := fmt.Sprintf(
			`%s: %s`,
			string(apis.Failed),
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

	var newStatus pipelineshub.RunConfigurationStatus
	state := apis.Succeeded
	message := ""

	if resolvedParameters, err := runConfiguration.Spec.Run.ResolveParameters(
		runConfiguration.Status.Dependencies,
	); err == nil {
		if hasChanged, err := r.syncWithRuns(ctx, runConfiguration); hasChanged || err != nil {
			return ctrl.Result{}, err
		}

		state, message, err = r.syncStatus(ctx, runConfiguration, resolvedParameters)
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
	runConfiguration *pipelineshub.RunConfiguration,
	indicator *triggers.Indicator,
) error {
	runs, err := findOwnedRuns(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return err
	}

	desiredRun, err := r.constructRunForRunConfiguration(runConfiguration)
	if err != nil {
		return err
	}

	if runExists := slices.ContainsFunc(runs, func(run pipelineshub.Run) bool {
		return string(run.ComputeHash()) == string(desiredRun.ComputeHash())
	}); runExists {
		return nil
	}

	if indicator != nil {
		if desiredRun.Labels == nil {
			desiredRun.Labels = map[string]string{}
		}
		desiredRun.Labels = lo.Assign(desiredRun.Labels, indicator.AsK8sLabels())
	}

	return r.EC.Client.Create(ctx, desiredRun)
}

func (r *RunConfigurationReconciler) updateRcTriggers(
	runConfiguration pipelineshub.RunConfiguration,
) pipelineshub.RunConfigurationStatus {
	newStatus := runConfiguration.Status

	if slices.Contains(runConfiguration.Spec.Triggers.OnChange, pipelineshub.OnChangeTypes.Pipeline) {
		newStatus.Triggers.Pipeline.Version = runConfiguration.Status.Dependencies.Pipeline.Version
	}

	if slices.Contains(runConfiguration.Spec.Triggers.OnChange, pipelineshub.OnChangeTypes.RunSpec) {
		newStatus.Triggers.RunSpec.Version = runConfiguration.Spec.Run.ComputeVersion()
	}

	newStatus.Triggers.RunConfigurations = lo.Associate(
		runConfiguration.Spec.Triggers.RunConfigurations,
		func(rcName common.NamespacedName) (string, pipelineshub.TriggeredRunReference) {
			rcNamespacedName, err := rcName.String()
			if err != nil {
				return "", pipelineshub.TriggeredRunReference{}
			}

			triggeredRun := pipelineshub.TriggeredRunReference{
				ProviderId: runConfiguration.Status.Dependencies.RunConfigurations[rcNamespacedName].ProviderId,
			}
			return rcNamespacedName, triggeredRun
		},
	)

	return newStatus
}

func (r *RunConfigurationReconciler) syncWithRuns(
	ctx context.Context,
	runConfiguration *pipelineshub.RunConfiguration,
) (bool, error) {
	oldStatus := runConfiguration.Status
	runConfiguration.Status = r.updateRcTriggers(*runConfiguration)

	triggersEqual := runConfiguration.Status.Triggers.Equals(oldStatus.Triggers)
	pipelineVersionEqual := runConfiguration.Status.Triggers.Pipeline.Version == oldStatus.Triggers.Pipeline.Version

	if triggersEqual && pipelineVersionEqual {
		return false, nil
	}

	triggerIndication := r.IdentifyRunTriggerReason(runConfiguration, oldStatus)

	if err := r.triggerUntriggeredRuns(ctx, runConfiguration, triggerIndication); err != nil {
		return false, err
	}

	return true, r.EC.Client.Status().Update(ctx, runConfiguration)
}

func (r *RunConfigurationReconciler) IdentifyRunTriggerReason(runConfiguration *pipelineshub.RunConfiguration, oldStatus pipelineshub.RunConfigurationStatus) *triggers.Indicator {

	if runConfiguration.Status.Triggers.RunSpec.Version != "" && runConfiguration.Status.Triggers.RunSpec.Version != oldStatus.Triggers.RunSpec.Version {
		return &triggers.Indicator{
			Type:            triggers.OnChangeRunSpec,
			Source:          runConfiguration.Name,
			SourceNamespace: runConfiguration.Namespace,
		}
	}

	if runConfiguration.Status.Triggers.Pipeline.Version != "" && runConfiguration.Status.Triggers.Pipeline.Version != oldStatus.Triggers.Pipeline.Version {
		return &triggers.Indicator{
			Type:            triggers.OnChangePipeline,
			Source:          runConfiguration.Spec.Run.Pipeline.Name,
			SourceNamespace: runConfiguration.Namespace,
		}
	}

	if runConfiguration.Status.Triggers.RunConfigurations != nil {
		differentRCs := GetDiffering(runConfiguration.Status.Triggers.RunConfigurations, oldStatus.Triggers.RunConfigurations)
		firstDifferentRc, hasDifferentRC := lo.First(differentRCs)

		if hasDifferentRC {
			rcNamespaceName, err := common.NamespacedNameFromString(firstDifferentRc)
			if err != nil {
				return nil
			}

			return &triggers.Indicator{
				Type:            triggers.RunConfiguration,
				Source:          rcNamespaceName.Name,
				SourceNamespace: rcNamespaceName.Namespace,
			}
		}
	}

	return nil
}

// GetDiffering returns a slice of keys that are present in both input maps `a` and `b`
// but have different values associated with them. The function only considers keys
// that exist in both maps and compares their corresponding values.
func GetDiffering[K, V comparable](a, b map[K]V) []K {
	differs := []K{}
	for k, vA := range a {
		if vB, ok := b[k]; ok && vA != vB {
			differs = append(differs, k)
		}
	}
	return differs
}

func (r *RunConfigurationReconciler) syncStatus(
	ctx context.Context,
	runConfiguration *pipelineshub.RunConfiguration,
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

	missingSchedules := apis.SliceDiff(desiredSchedules, dependentSchedules, compareRunSchedules)
	excessSchedules := apis.SliceDiff(dependentSchedules, desiredSchedules, compareRunSchedules)
	excessSchedulesNotMarkedForDeletion := lo.Filter(
		excessSchedules, func(schedule pipelineshub.RunSchedule, _ int) bool {
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
	referencingRunConfigurations := &pipelineshub.RunConfigurationList{}
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
	referencingRunConfigurations := &pipelineshub.RunConfigurationList{}
	rcNamespacedName, err := common.NamespacedName{
		Name:      runConfiguration.GetName(),
		Namespace: runConfiguration.GetNamespace(),
	}.String()
	if err != nil {
		return []reconcile.Request{}
	}

	rcRefListOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rcRefField, rcNamespacedName),
	}

	err = r.EC.Client.Cached.List(ctx, referencingRunConfigurations, rcRefListOps)
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
	runConfiguration := &pipelineshub.RunConfiguration{}
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(runConfiguration).WithOptions(controller.Options{
		RateLimiter: controllerconfigutil.RateLimiter,
	})

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

	return controllerBuilder.Owns(&pipelineshub.RunSchedule{}).Complete(r)
}

func findOwnedRunSchedules(
	ctx context.Context,
	cli client.Reader,
	runConfiguration *pipelineshub.RunConfiguration,
) ([]pipelineshub.RunSchedule, error) {
	ownedRunSchedulesList := &pipelineshub.RunScheduleList{}
	if err := cli.List(
		ctx,
		ownedRunSchedulesList,
		client.InNamespace(runConfiguration.Namespace),
	); err != nil {
		return nil, err
	}

	var ownedSchedules []pipelineshub.RunSchedule
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
	runConfiguration *pipelineshub.RunConfiguration,
) ([]pipelineshub.Run, error) {
	ownedRunsList := &pipelineshub.RunList{}
	if err := cli.List(ctx, ownedRunsList, client.InNamespace(runConfiguration.Namespace)); err != nil {
		return nil, err
	}

	var ownedRuns []pipelineshub.Run
	for _, run := range ownedRunsList.Items {
		if metav1.IsControlledBy(&run, runConfiguration) {
			ownedRuns = append(ownedRuns, run)
		}
	}

	return ownedRuns, nil
}

func (r *RunConfigurationReconciler) constructRunForRunConfiguration(
	runConfiguration *pipelineshub.RunConfiguration,
) (*pipelineshub.Run, error) {
	spec := runConfiguration.Spec.Run
	spec.Pipeline.Version = runConfiguration.Status.Dependencies.Pipeline.Version

	run := pipelineshub.Run{
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
	runConfiguration *pipelineshub.RunConfiguration,
	resolvedParameters []apis.NamedValue,
) ([]pipelineshub.RunSchedule, error) {
	var schedules []pipelineshub.RunSchedule

	for _, schedule := range runConfiguration.Spec.Triggers.Schedules {
		runSchedule := pipelineshub.RunSchedule{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: runConfiguration.Name + "-",
				Namespace:    runConfiguration.Namespace,
			},
			Spec: pipelineshub.RunScheduleSpec{
				Provider: runConfiguration.Spec.Run.Provider,
				Pipeline: pipelineshub.PipelineIdentifier{
					Name:    runConfiguration.Spec.Run.Pipeline.Name,
					Version: runConfiguration.Status.Dependencies.Pipeline.Version,
				},
				Parameters:     resolvedParameters,
				ExperimentName: runConfiguration.Spec.Run.ExperimentName,
				Artifacts:      runConfiguration.Spec.Run.Artifacts,
				Schedule:       schedule,
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

func aggregateState(dependencies []pipelineshub.RunSchedule) (aggState apis.SynchronizationState, message string) {
	aggState = apis.Succeeded

	for _, dependency := range dependencies {
		state := dependency.Status.Conditions.GetSyncStateFromReason()
		if state == apis.Failed {
			aggState = apis.Failed
			message = dependency.Status.Conditions.SynchronizationSucceeded().Message
		} else if state != apis.Succeeded {
			aggState = apis.Updating
			message = "Waiting for all dependant runschedules to be in a state of Succeeded"
			return
		}
	}

	return aggState, message
}

func compareRunSchedules(a, b pipelineshub.RunSchedule) bool {
	return string(a.ComputeHash()) == string(b.ComputeHash())
}
