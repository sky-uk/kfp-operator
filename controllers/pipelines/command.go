package pipelines

import (
	"context"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var EventTypes = struct {
	Normal  string
	Warning string
}{
	Warning: "Warning",
	Normal:  "Normal",
}

var EventReasons = struct {
	Syncing    string
	Synced     string
	SyncFailed string
}{
	Syncing:    "Syncing",
	Synced:     "Synced",
	SyncFailed: "SyncFailed",
}

type K8sExecutionContext struct {
	Client             controllers.OptInClient
	Recorder           record.EventRecorder
	Scheme             *runtime.Scheme
	WorkflowRepository WorkflowRepository
}

type Command interface {
	execute(context.Context, K8sExecutionContext, pipelineshub.Resource) error
}

func alwaysSetObservedGeneration(ctx context.Context, commands []Command, resource pipelineshub.Resource, time metav1.Time) []Command {
	currentGeneration := resource.GetGeneration()
	if currentGeneration == resource.GetStatus().ObservedGeneration {
		return commands
	}

	logger := log.FromContext(ctx)
	setStatusExists := false
	var modifiedCommands []Command

	for _, command := range commands {
		setStatus, ok := command.(SetStatus)

		if ok {
			if setStatusExists {
				logger.Info("attempting to set status more than once in the same reconciliation, this is likely to cause inconsistencies")
			}

			message := setStatus.Status.Conditions.SynchronizationSucceeded().Message

			setStatusExists = true
			setStatus.Status.ObservedGeneration = currentGeneration
			setStatus.Status.Conditions.SetObservedGeneration(
				message,
				time,
				currentGeneration,
			)

			modifiedCommands = append(modifiedCommands, setStatus)
		} else {
			modifiedCommands = append(modifiedCommands, command)
		}
	}

	if !setStatusExists {
		newStatus := resource.GetStatus()
		newStatus.ObservedGeneration = currentGeneration
		modifiedCommands = append(modifiedCommands, SetStatus{Status: newStatus})
	}

	return modifiedCommands
}

type SetStatus struct {
	Status pipelineshub.Status
}

func From(status pipelineshub.Status) *SetStatus {
	return &SetStatus{
		Status: status,
	}
}

func NewSetStatus() *SetStatus {
	return &SetStatus{}
}

func (sps *SetStatus) WithSyncStateCondition(state apis.SynchronizationState, time metav1.Time, message string) *SetStatus {
	condition := metav1.Condition{
		Type:               apis.ConditionTypes.SynchronizationSucceeded,
		Status:             apis.ConditionStatusForSynchronizationState(state),
		ObservedGeneration: sps.Status.ObservedGeneration,
		LastTransitionTime: time,
		Reason:             string(state),
		Message:            message,
	}

	if sps.Status.Conditions == nil {
		sps.Status.Conditions = apis.Conditions{condition}
	} else {
		sps.Status.Conditions = sps.Status.Conditions.MergeIntoConditions(condition)
	}

	return sps
}

func (sps *SetStatus) WithVersion(version string) *SetStatus {
	sps.Status.Version = version

	return sps
}

func (sps *SetStatus) WithProvider(providerId pipelineshub.ProviderAndId) *SetStatus {
	sps.Status.Provider = providerId

	return sps
}

func eventMessage(sps SetStatus) (message string) {
	message = fmt.Sprintf(`%s [version: "%s"]`, string(sps.Status.Conditions.GetSyncStateFromReason()), sps.Status.Version)

	conditionMessage := sps.Status.Conditions.SynchronizationSucceeded().Message

	if conditionMessage != "" {
		message = fmt.Sprintf("%s: %s", message, conditionMessage)
	}

	return
}

func eventType(sps SetStatus) string {
	if sps.Status.Conditions.GetSyncStateFromReason() == apis.Failed {
		return EventTypes.Warning
	} else {
		return EventTypes.Normal
	}
}

func eventReason(sps SetStatus) string {
	switch sps.Status.Conditions.GetSyncStateFromReason() {
	case apis.Succeeded, apis.Deleted:
		return EventReasons.Synced
	case apis.Failed:
		return EventReasons.SyncFailed
	default:
		return EventReasons.Syncing
	}
}

func (sps SetStatus) execute(
	ctx context.Context,
	ec K8sExecutionContext,
	resource pipelineshub.Resource,
) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info(
		"setting pipeline status",
		logkeys.OldStatus,
		resource.GetStatus(),
		logkeys.NewStatus,
		sps.Status,
	)

	resource.SetStatus(sps.Status)

	err := ec.Client.Status().Update(ctx, resource)

	if err == nil {
		ec.Recorder.Event(
			resource,
			eventType(sps),
			eventReason(sps),
			eventMessage(sps),
		)
	}

	return err
}

type CreateWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateWorkflow) execute(
	ctx context.Context,
	ec K8sExecutionContext,
	resource pipelineshub.Resource,
) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info(
		"creating child workflow",
		logkeys.Workflow,
		cw.Workflow,
	)

	if err := ec.WorkflowRepository.CreateWorkflowForResource(
		ctx,
		&cw.Workflow,
		resource,
	); err != nil {
		return err
	}

	return nil
}

type MarkWorkflowsAsProcessed struct {
	Workflows []argo.Workflow
}

func (dw MarkWorkflowsAsProcessed) execute(ctx context.Context, ec K8sExecutionContext, _ pipelineshub.Resource) error {
	for i := range dw.Workflows {
		workflow := &dw.Workflows[i]
		if err := ec.WorkflowRepository.MarkWorkflowAsProcessed(ctx, workflow); err != nil {
			return err
		}
	}

	return nil
}

type AcquireResource struct {
}

func (ap AcquireResource) execute(ctx context.Context, ec K8sExecutionContext, resource pipelineshub.Resource) error {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(resource, finalizerName) {
		logger.V(2).Info("adding finalizer")
		controllerutil.AddFinalizer(resource, finalizerName)

		return ec.Client.Update(ctx, resource)
	}

	return nil
}

type ReleaseResource struct {
}

func (rp ReleaseResource) execute(ctx context.Context, ec K8sExecutionContext, resource pipelineshub.Resource) error {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(resource, finalizerName) {
		logger.V(2).Info("removing finalizer")
		controllerutil.RemoveFinalizer(resource, finalizerName)
		return ec.Client.Update(ctx, resource)
	}

	return nil
}
