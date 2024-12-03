package pipelines

import (
	"context"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logging"
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
	Execute(context.Context, K8sExecutionContext, pipelinesv1.Resource) error
}

func AlwaysSetObservedGeneration(ctx context.Context, commands []Command, resource pipelinesv1.Resource) []Command {
	currentGeneration := resource.GetGeneration()
	if currentGeneration == resource.GetStatus().ObservedGeneration {
		return commands
	}

	logger := log.FromContext(ctx)
	setStatusExists := false
	var modifiedCommands []Command

	for _, command := range commands {
		setStatus, isSetStatus := command.(SetStatus)

		if isSetStatus {
			if setStatusExists {
				logger.Info("attempting to set status more than once in the same reconciliation, this is likely to cause inconsistencies")
			}

			setStatusExists = true
			setStatus.Status.ObservedGeneration = currentGeneration
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
	Message string
	Status  pipelinesv1.Status
}

func From(status pipelinesv1.Status) *SetStatus {
	return &SetStatus{
		Status: status,
	}
}

func NewSetStatus() *SetStatus {
	return &SetStatus{}
}

func (sps *SetStatus) WithSynchronizationState(state apis.SynchronizationState) *SetStatus {
	sps.Status.SynchronizationState = state

	return sps
}

func (sps *SetStatus) WithVersion(version string) *SetStatus {
	sps.Status.Version = version

	return sps
}

func (sps *SetStatus) WithProvider(providerId pipelinesv1.ProviderAndId) *SetStatus {
	sps.Status.Provider = providerId

	return sps
}

func (sps *SetStatus) WithMessage(message string) *SetStatus {
	sps.Message = message

	return sps
}

func eventMessage(sps SetStatus) (message string) {
	message = fmt.Sprintf(`%s [version: "%s"]`, string(sps.Status.SynchronizationState), sps.Status.Version)

	if sps.Message != "" {
		message = fmt.Sprintf("%s: %s", message, sps.Message)
	}

	return
}

func eventType(sps SetStatus) string {
	if sps.Status.SynchronizationState == apis.Failed {
		return EventTypes.Warning
	} else {
		return EventTypes.Normal
	}
}

func eventReason(sps SetStatus) string {
	switch sps.Status.SynchronizationState {
	case apis.Succeeded, apis.Deleted:
		return EventReasons.Synced
	case apis.Failed:
		return EventReasons.SyncFailed
	default:
		return EventReasons.Syncing
	}
}

func (sps SetStatus) statusWithCondition() pipelinesv1.Status {
	sps.Status.Conditions = sps.Status.Conditions.MergeIntoConditions(metav1.Condition{
		LastTransitionTime: metav1.Now(),
		Message:            sps.Message,
		ObservedGeneration: sps.Status.ObservedGeneration,
		Type:               pipelinesv1.ConditionTypes.SynchronizationSucceeded,
		Status:             pipelinesv1.ConditionStatusForSynchronizationState(sps.Status.SynchronizationState),
		Reason:             string(sps.Status.SynchronizationState),
	})

	return sps.Status
}

func (sps SetStatus) Execute(ctx context.Context, ec K8sExecutionContext, resource pipelinesv1.Resource) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("setting pipeline status", logging.LogKeys.OldStatus, resource.GetStatus(), logging.LogKeys.NewStatus, sps.Status)

	resource.SetStatus(sps.statusWithCondition())

	err := ec.Client.Status().Update(ctx, resource)

	if err == nil {
		ec.Recorder.Event(resource, eventType(sps), eventReason(sps), eventMessage(sps))
	}

	return err
}

type CreateWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateWorkflow) Execute(ctx context.Context, ec K8sExecutionContext, resource pipelinesv1.Resource) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("creating child workflow", logging.LogKeys.Workflow, cw.Workflow)

	if err := ec.WorkflowRepository.CreateWorkflowForResource(ctx, &cw.Workflow, resource); err != nil {
		return err
	}

	return nil
}

type MarkWorkflowsAsProcessed struct {
	Workflows []argo.Workflow
}

func (dw MarkWorkflowsAsProcessed) Execute(ctx context.Context, ec K8sExecutionContext, _ pipelinesv1.Resource) error {
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

func (ap AcquireResource) Execute(ctx context.Context, ec K8sExecutionContext, resource pipelinesv1.Resource) error {
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

func (rp ReleaseResource) Execute(ctx context.Context, ec K8sExecutionContext, resource pipelinesv1.Resource) error {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(resource, finalizerName) {
		logger.V(2).Info("removing finalizer")
		controllerutil.RemoveFinalizer(resource, finalizerName)
		return ec.Client.Update(ctx, resource)
	}

	return nil
}
