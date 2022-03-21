package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
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
	Client   controllers.OptInClient
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

type Command interface {
	execute(context.Context, K8sExecutionContext, Resource) error
}

func alwaysSetObservedGeneration(ctx context.Context, commands []Command, resource Resource) []Command {
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

func (sps *SetStatus) WithSynchronizationState(state pipelinesv1.SynchronizationState) *SetStatus {
	sps.Status.SynchronizationState = state

	return sps
}

func (sps *SetStatus) WithVersion(version string) *SetStatus {
	sps.Status.Version = version

	return sps
}

func (sps *SetStatus) WithKfpId(kfpId string) *SetStatus {
	sps.Status.KfpId = kfpId

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
	if sps.Status.SynchronizationState == pipelinesv1.Failed {
		return EventTypes.Warning
	} else {
		return EventTypes.Normal
	}
}

func eventReason(sps SetStatus) string {
	switch sps.Status.SynchronizationState {
	case pipelinesv1.Succeeded, pipelinesv1.Deleted:
		return EventReasons.Synced
	case pipelinesv1.Failed:
		return EventReasons.SyncFailed
	default:
		return EventReasons.Syncing
	}
}

func (sps SetStatus) execute(ctx context.Context, ec K8sExecutionContext, resource Resource) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("setting pipeline status", LogKeys.OldStatus, resource.GetStatus(), LogKeys.NewStatus, sps.Status)

	resource.SetStatus(sps.Status)

	err := ec.Client.Status().Update(ctx, resource)

	if err == nil {
		ec.Recorder.Event(resource, eventType(sps), eventReason(sps), eventMessage(sps))
	}

	return err
}

type CreateWorkflow struct {
	Workflow argo.Workflow
}

func (cw CreateWorkflow) execute(ctx context.Context, ec K8sExecutionContext, resource Resource) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("creating child workflow", LogKeys.Workflow, cw.Workflow)

	if err := ctrl.SetControllerReference(resource, &cw.Workflow, ec.Scheme); err != nil {
		return err
	}

	if err := ec.Client.Create(ctx, &cw.Workflow); err != nil {
		return err
	}

	return nil
}

type DeleteWorkflows struct {
	Workflows []argo.Workflow
}

func (dw DeleteWorkflows) execute(ctx context.Context, ec K8sExecutionContext, resource Resource) error {
	logger := log.FromContext(ctx)

	for i := range dw.Workflows {
		workflow := &dw.Workflows[i]
		workflowDebugOptions := pipelinesv1.DebugOptionsFromAnnotations(ctx, workflow.ObjectMeta.Annotations)
		if !workflowDebugOptions.KeepWorkflows {
			logger.V(1).Info("deleting child workflow", LogKeys.Workflow, workflow)
			if err := ec.Client.Delete(ctx, workflow); err != nil {
				return err
			}
		} else {
			logger.V(2).Info("keeping child workflow", LogKeys.Workflow, workflow)
		}
	}

	return nil
}

type AcquireResource struct {
}

func (ap AcquireResource) execute(ctx context.Context, ec K8sExecutionContext, resource Resource) error {
	logger := log.FromContext(ctx)

	if !containsString(resource.GetFinalizers(), finalizerName) {
		logger.V(2).Info("adding finalizer")
		resource.SetFinalizers(append(resource.GetFinalizers(), finalizerName))

		return ec.Client.Update(ctx, resource)
	}

	return nil
}

type ReleaseResource struct {
}

func (rp ReleaseResource) execute(ctx context.Context, ec K8sExecutionContext, resource Resource) error {
	logger := log.FromContext(ctx)

	if containsString(resource.GetFinalizers(), finalizerName) {
		logger.V(2).Info("removing finalizer")
		resource.SetFinalizers(removeString(resource.GetFinalizers(), finalizerName))
		return ec.Client.Update(ctx, resource)
	}

	return nil
}
