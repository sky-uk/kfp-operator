package pipelines

import (
	"context"
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	rcTriggersField = ".spec.triggers.runConfigurations"
)

// RunConfigurationReconciler reconciles a RunConfiguration object
type RunConfigurationReconciler struct {
	DependingOnPipelineReconciler[*pipelinesv1.RunConfiguration]
	DependingOnRunConfigurationReconciler[*pipelinesv1.RunConfiguration]
	EC     K8sExecutionContext
	Scheme *runtime.Scheme
	Config config.Configuration
}

func NewRunConfigurationReconciler(ec K8sExecutionContext, scheme *runtime.Scheme, config config.Configuration) *RunConfigurationReconciler {
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

	// For migration from v1alpha4. Remove afterwards.
	if controllerutil.ContainsFinalizer(runConfiguration, finalizerName) {
		logger.V(3).Info("Removing finalizer", "resource", runConfiguration)
		controllerutil.RemoveFinalizer(runConfiguration, finalizerName)
		return ctrl.Result{}, r.EC.Client.Update(ctx, runConfiguration)
	}

	if runConfiguration.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	desiredProvider := desiredProvider(runConfiguration, r.Config)
	if runConfiguration.Status.Provider != "" && desiredProvider != runConfiguration.Status.Provider {
		//TODO: refactor to use Commands and introduce a StateHandler
		runConfiguration.Status.SynchronizationState = apis.Failed

		message := fmt.Sprintf(`%s: %s`, string(runConfiguration.Status.SynchronizationState), StateHandlerConstants.ProviderChangedError)
		r.EC.Recorder.Event(runConfiguration, EventTypes.Warning, EventReasons.SyncFailed, message)

		return ctrl.Result{}, r.EC.Client.Status().Update(ctx, runConfiguration)
	}

	if hasChanged, err := r.handleObservedPipelineVersion(ctx, runConfiguration.Spec.Run.Pipeline, runConfiguration); hasChanged || err != nil {
		return ctrl.Result{}, err
	}

	if hasChanged, err := r.handleDependentRuns(ctx, runConfiguration); hasChanged || err != nil {
		return ctrl.Result{}, err
	}

	hasUntriggeredPipelineDependency := runConfiguration.Status.ObservedPipelineVersion != runConfiguration.Status.TriggeredPipelineVersion && runConfiguration.Spec.Triggers.OnChange != nil
	hasUntriggeredRcDependency := pipelines.Exists(runConfiguration.Spec.Triggers.RunConfigurations, func(rcName string) bool {
		return runConfiguration.Status.Triggers.RunConfigurations[rcName].ProviderId == "" || runConfiguration.Status.Triggers.RunConfigurations[rcName].ProviderId != runConfiguration.Status.Dependencies.RunConfigurations[rcName].ProviderId
	})

	if hasUntriggeredPipelineDependency || hasUntriggeredRcDependency {
		return ctrl.Result{}, r.syncWithRuns(ctx, desiredProvider, runConfiguration)
	}

	newStatus := runConfiguration.Status
	newStatus.ObservedGeneration = runConfiguration.GetGeneration()

	newSynchronizationState, err := r.syncWithRunSchedules(ctx, desiredProvider, runConfiguration)
	if err != nil {
		return ctrl.Result{}, err
	}

	newStatus.SynchronizationState = newSynchronizationState
	newStatus.Provider = desiredProvider

	if !reflect.DeepEqual(newStatus, runConfiguration.Status) {
		runConfiguration.Status = newStatus
		if err := r.EC.Client.Status().Update(ctx, runConfiguration); err != nil {
			return ctrl.Result{}, err
		}
	}

	duration := time.Now().Sub(startTime)
	logger.V(2).Info("reconciliation ended", LogKeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *RunConfigurationReconciler) syncWithRuns(ctx context.Context, provider string, runConfiguration *pipelinesv1.RunConfiguration) error {
	runs, err := findOwnedRuns(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return err
	}

	desiredRun, err := r.constructRunForRunConfiguration(provider, runConfiguration)
	if err != nil {
		return err
	}

	runExists := pipelines.Exists(runs, func(run pipelinesv1.Run) bool {
		return string(run.ComputeHash()) == string(desiredRun.ComputeHash())
	})

	if !runExists {
		if err := r.EC.Client.Create(ctx, desiredRun); err != nil {
			return err
		}
	}

	runConfiguration.Status.TriggeredPipelineVersion = runConfiguration.Status.ObservedPipelineVersion
	for _, rcName := range runConfiguration.Spec.Triggers.RunConfigurations {
		if runConfiguration.Status.Triggers.RunConfigurations == nil {
			runConfiguration.Status.Triggers.RunConfigurations = make(map[string]pipelinesv1.TriggeredRunReference)
		}
		triggeredRc := runConfiguration.Status.Triggers.RunConfigurations[rcName]
		triggeredRc.ProviderId = runConfiguration.Status.Dependencies.RunConfigurations[rcName].ProviderId
		runConfiguration.Status.Triggers.RunConfigurations[rcName] = triggeredRc
	}

	return r.EC.Client.Status().Update(ctx, runConfiguration)
}

func (r *RunConfigurationReconciler) syncWithRunSchedules(ctx context.Context, provider string, runConfiguration *pipelinesv1.RunConfiguration) (state apis.SynchronizationState, err error) {
	state = runConfiguration.Status.SynchronizationState

	desiredSchedules, err := r.constructRunSchedulesForTriggers(provider, runConfiguration)
	if err != nil {
		return
	}

	dependentSchedules, err := findOwnedRunSchedules(ctx, r.EC.Client.NonCached, runConfiguration)
	if err != nil {
		return
	}

	missingSchedules := pipelines.SliceDiff(desiredSchedules, dependentSchedules, compareRunSchedules)
	excessSchedules := pipelines.SliceDiff(dependentSchedules, desiredSchedules, compareRunSchedules)
	excessSchedulesNotMarkedForDeletion := pipelines.Filter(excessSchedules, func(schedule pipelinesv1.RunSchedule) bool {
		return schedule.DeletionTimestamp == nil
	})

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

func (r *RunConfigurationReconciler) reconciliationRequestsForRunConfiguration(runConfiguration client.Object) []reconcile.Request {
	referencingRunConfigurations := &pipelinesv1.RunConfigurationList{}
	rcRefListOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rcRefField, runConfiguration.GetName()),
		Namespace:     runConfiguration.GetNamespace(),
	}
	rcTriggersListOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(rcTriggersField, runConfiguration.GetName()),
		Namespace:     runConfiguration.GetNamespace(),
	}

	err := r.EC.Client.Cached.List(context.TODO(), referencingRunConfigurations, rcRefListOps, rcTriggersListOps)
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

	controllerBuilder, err := r.DependingOnPipelineReconciler.setupWithManager(mgr, controllerBuilder, runConfiguration, r.reconciliationRequestsForPipeline)
	if err != nil {
		return err
	}
	controllerBuilder, err = r.DependingOnRunConfigurationReconciler.setupWithManager(mgr, controllerBuilder, runConfiguration, r.reconciliationRequestsForRunConfiguration)
	if err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), runConfiguration, rcRefField, func(rawObj client.Object) []string {
		return pipelines.Map(rawObj.(*pipelinesv1.RunConfiguration).GetRCRuntimeParameters(), func(r pipelinesv1.RunConfigurationRef) string {
			return r.Name
		})
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), runConfiguration, rcTriggersField, func(rawObj client.Object) []string {
		return rawObj.(*pipelinesv1.RunConfiguration).Spec.Triggers.RunConfigurations
	}); err != nil {
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

func findOwnedRuns(ctx context.Context, cli client.Reader, runConfiguration *pipelinesv1.RunConfiguration) ([]pipelinesv1.Run, error) {
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

func (r *RunConfigurationReconciler) constructRunForRunConfiguration(provider string, runConfiguration *pipelinesv1.RunConfiguration) (*pipelinesv1.Run, error) {
	run := pipelinesv1.Run{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: runConfiguration.Name + "-",
			Namespace:    runConfiguration.Namespace,
		},
		Spec: pipelinesv1.RunSpec{
			Pipeline: pipelinesv1.PipelineIdentifier{
				Name:    runConfiguration.Spec.Run.Pipeline.Name,
				Version: runConfiguration.Status.ObservedPipelineVersion,
			},
			RuntimeParameters: runConfiguration.Spec.Run.RuntimeParameters,
			ExperimentName:    runConfiguration.Spec.Run.ExperimentName,
		},
	}

	if err := controllerutil.SetControllerReference(runConfiguration, &run, r.Scheme); err != nil {
		return nil, err
	}
	metav1.SetMetaDataAnnotation(&run.ObjectMeta, apis.ResourceAnnotations.Provider, provider)

	return &run, nil
}

func (r *RunConfigurationReconciler) constructRunSchedulesForTriggers(provider string, runConfiguration *pipelinesv1.RunConfiguration) ([]pipelinesv1.RunSchedule, error) {
	var schedules []pipelinesv1.RunSchedule

	runtimeParameters, err := runConfiguration.Spec.Run.ResolveRuntimeParameters(runConfiguration.Status.Dependencies)
	if err != nil {
		return nil, err
	}

	for _, schedule := range runConfiguration.Spec.Triggers.Schedules {
		runSchedule := pipelinesv1.RunSchedule{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: runConfiguration.Name + "-",
				Namespace:    runConfiguration.Namespace,
			},
			Spec: pipelinesv1.RunScheduleSpec{
				Pipeline: pipelinesv1.PipelineIdentifier{
					Name:    runConfiguration.Spec.Run.Pipeline.Name,
					Version: runConfiguration.Status.ObservedPipelineVersion,
				},
				RuntimeParameters: runtimeParameters,
				ExperimentName:    runConfiguration.Spec.Run.ExperimentName,
				Artifacts:         runConfiguration.Spec.Run.Artifacts,
				Schedule:          schedule,
			},
		}
		if err := controllerutil.SetControllerReference(runConfiguration, &runSchedule, r.Scheme); err != nil {
			return nil, err
		}
		metav1.SetMetaDataAnnotation(&runSchedule.ObjectMeta, apis.ResourceAnnotations.Provider, provider)
		schedules = append(schedules, runSchedule)
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
