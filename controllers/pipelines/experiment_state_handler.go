package pipelines

import (
	"context"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ExperimentStateHandler struct {
	WorkflowFactory    ExperimentWorkflowFactory
	WorkflowRepository WorkflowRepository
}

func (st *ExperimentStateHandler) StateTransition(ctx context.Context, experiment *pipelinesv1.Experiment) (commands []ExperimentCommand) {
	logger := log.FromContext(ctx)
	logger.Info("state transition start")

	switch experiment.Status.SynchronizationState {
	case pipelinesv1.Creating:
		commands = st.onCreating(ctx, experiment,
			st.WorkflowRepository.GetByOperation(ctx,
				ExperimentWorkflowConstants.CreateOperationLabel,
				experiment.NamespacedName(),
				ExperimentWorkflowConstants.ExperimentNameLabelKey))
	case pipelinesv1.Succeeded, pipelinesv1.Failed:
		if !experiment.ObjectMeta.DeletionTimestamp.IsZero() {
			commands = st.onDelete(ctx, experiment)
		} else {
			commands = st.onSucceededOrFailed(ctx, experiment)
		}
	case pipelinesv1.Updating:
		commands = st.onUpdating(ctx, experiment,
			st.WorkflowRepository.GetByOperation(ctx,
				ExperimentWorkflowConstants.UpdateOperationLabel,
				experiment.NamespacedName(),
				ExperimentWorkflowConstants.ExperimentNameLabelKey))
	case pipelinesv1.Deleting:
		commands = st.onDeleting(ctx, experiment,
			st.WorkflowRepository.GetByOperation(ctx,
				ExperimentWorkflowConstants.DeleteOperationLabel,
				experiment.NamespacedName(),
				ExperimentWorkflowConstants.ExperimentNameLabelKey))
	case pipelinesv1.Deleted:
	default:
		commands = st.onUnknown(ctx, experiment)
	}

	if experiment.Status.SynchronizationState == pipelinesv1.Deleted {
		commands = append([]ExperimentCommand{ReleaseExperiment{}}, commands...)
	} else {
		commands = append([]ExperimentCommand{AcquireExperiment{}}, commands...)
	}

	return
}

func (st *ExperimentStateHandler) onUnknown(ctx context.Context, experiment *pipelinesv1.Experiment) []ExperimentCommand {
	logger := log.FromContext(ctx)

	newRunconfigurationVersion := experiment.Spec.ComputeVersion()

	if experiment.Status.KfpId != "" {
		logger.Info("empty state but KfpId already exists, updating run configuration")
		workflow, err := st.WorkflowFactory.ConstructUpdateWorkflow(ctx, experiment)

		if err != nil {
			logger.Error(err, "error constructing update workflow, failing experiment")

			return []ExperimentCommand{
				SetExperimentStatus{
					Status: pipelinesv1.Status{
						Version:              experiment.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

		return []ExperimentCommand{
			SetExperimentStatus{
				Status: pipelinesv1.Status{
					KfpId:                experiment.Status.KfpId,
					Version:              newRunconfigurationVersion,
					SynchronizationState: pipelinesv1.Updating,
				},
			},
			CreateExperimentWorkflow{Workflow: *workflow},
		}
	}

	logger.Info("empty state, creating run configuration")
	workflow, err := st.WorkflowFactory.ConstructCreationWorkflow(ctx, experiment)

	if err != nil {
		logger.Error(err, "error constructing creation workflow, failing experiment")

		return []ExperimentCommand{
			SetExperimentStatus{
				Status: pipelinesv1.Status{
					Version:              experiment.Status.Version,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []ExperimentCommand{
		SetExperimentStatus{
			Status: pipelinesv1.Status{
				Version:              newRunconfigurationVersion,
				SynchronizationState: pipelinesv1.Creating,
			},
		},
		CreateExperimentWorkflow{Workflow: *workflow},
	}
}

func (st ExperimentStateHandler) onDelete(ctx context.Context, experiment *pipelinesv1.Experiment) []ExperimentCommand {
	logger := log.FromContext(ctx)
	logger.Info("deletion requested, deleting")

	if experiment.Status.KfpId == "" {
		return []ExperimentCommand{
			SetExperimentStatus{
				Status: pipelinesv1.Status{
					KfpId:                experiment.Status.KfpId,
					Version:              experiment.Status.Version,
					SynchronizationState: pipelinesv1.Deleted,
				},
			},
		}
	}

	workflow, err := st.WorkflowFactory.ConstructDeletionWorkflow(ctx, experiment)

	if err != nil {
		logger.Error(err, "error constructing deletion workflow, failing experiment")

		return []ExperimentCommand{
			SetExperimentStatus{
				Status: pipelinesv1.Status{
					Version:              experiment.Status.Version,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	return []ExperimentCommand{
		SetExperimentStatus{
			Status: pipelinesv1.Status{
				KfpId:                experiment.Status.KfpId,
				Version:              experiment.Status.Version,
				SynchronizationState: pipelinesv1.Deleting,
			},
		},
		CreateExperimentWorkflow{Workflow: *workflow},
	}
}

func (st ExperimentStateHandler) onSucceededOrFailed(ctx context.Context, experiment *pipelinesv1.Experiment) []ExperimentCommand {
	logger := log.FromContext(ctx)
	newExperimentVersion := experiment.Spec.ComputeVersion()

	if experiment.Status.Version == newExperimentVersion {
		logger.V(2).Info("run configuration version has not changed")
		return []ExperimentCommand{}
	}

	var workflow *argo.Workflow
	var err error
	var targetState pipelinesv1.SynchronizationState

	if experiment.Status.KfpId == "" {
		logger.Info("no kfpId exists, creating")
		workflow, err = st.WorkflowFactory.ConstructCreationWorkflow(ctx, experiment)

		if err != nil {
			logger.Error(err, "error constructing creation workflow, failing experiment")

			return []ExperimentCommand{
				SetExperimentStatus{
					Status: pipelinesv1.Status{
						Version:              experiment.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

		targetState = pipelinesv1.Creating
	} else {
		logger.Info("kfpId exists, updating")
		workflow, err = st.WorkflowFactory.ConstructUpdateWorkflow(ctx, experiment)

		if err != nil {
			logger.Error(err, "error constructing update workflow, failing experiment")

			return []ExperimentCommand{
				SetExperimentStatus{
					Status: pipelinesv1.Status{
						Version:              experiment.Status.Version,
						SynchronizationState: pipelinesv1.Failed,
					},
				},
			}
		}

		targetState = pipelinesv1.Updating
	}

	return []ExperimentCommand{
		SetExperimentStatus{
			Status: pipelinesv1.Status{
				KfpId:                experiment.Status.KfpId,
				Version:              newExperimentVersion,
				SynchronizationState: targetState,
			},
		},
		CreateExperimentWorkflow{Workflow: *workflow},
	}
}

func (st ExperimentStateHandler) onUpdating(ctx context.Context, experiment *pipelinesv1.Experiment, updateWorkflows []argo.Workflow) []ExperimentCommand {
	logger := log.FromContext(ctx)

	if experiment.Status.Version == "" || experiment.Status.KfpId == "" {
		logger.Info("updating run configuration with empty version or kfpId, failing run configuration")
		return []ExperimentCommand{
			SetExperimentStatus{
				Status: pipelinesv1.Status{
					Version:              experiment.Status.Version,
					KfpId:                experiment.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(updateWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration update in progress")
		return []ExperimentCommand{}
	}

	statusAfterUpdating := func() (newStatus pipelinesv1.Status) {
		newStatus = *experiment.Status.DeepCopy()

		if succeeded == nil {
			if failed != nil {
				logger.Info("run configuration update failed")
			} else {
				logger.Info("run configuration creation progress unknown, failing run configuration")
			}

			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		idResult, _ := getWorkflowOutput(succeeded, ExperimentWorkflowConstants.ExperimentIdParameterName)

		if idResult == "" {
			logger.Info("could not retrieve kfpId, failing run configuration")
			newStatus.KfpId = ""
			newStatus.SynchronizationState = pipelinesv1.Failed
			return
		}

		logger.Info("run configuration update succeeded")
		newStatus.KfpId = idResult
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		return
	}

	return []ExperimentCommand{
		SetExperimentStatus{
			Status: statusAfterUpdating(),
		},
		DeleteExperimentWorkflows{
			Workflows: updateWorkflows,
		},
	}
}

func (st ExperimentStateHandler) onDeleting(ctx context.Context, experiment *pipelinesv1.Experiment, deletionWorkflows []argo.Workflow) []ExperimentCommand {
	logger := log.FromContext(ctx)

	inProgress, succeeded, failed := latestWorkflowByPhase(deletionWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration deletion in progress")
		return []ExperimentCommand{}
	}

	newStatus := experiment.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("run configuration deletion succeeded")
		newStatus.SynchronizationState = pipelinesv1.Deleted
	} else if failed != nil {
		logger.Info("run configuration deletion failed")
	} else {
		logger.Info("run configuration deletion progress unknown, failing run configuration")
	}

	return []ExperimentCommand{
		SetExperimentStatus{
			Status: *newStatus,
		},
		DeleteExperimentWorkflows{
			Workflows: deletionWorkflows,
		},
	}
}

func (st ExperimentStateHandler) onCreating(ctx context.Context, experiment *pipelinesv1.Experiment, creationWorkflows []argo.Workflow) []ExperimentCommand {
	logger := log.FromContext(ctx)

	if experiment.Status.Version == "" {
		logger.Info("creating run configuration with empty version, failing run configuration")
		return []ExperimentCommand{
			SetExperimentStatus{
				Status: pipelinesv1.Status{
					KfpId:                experiment.Status.KfpId,
					SynchronizationState: pipelinesv1.Failed,
				},
			},
		}
	}

	inProgress, succeeded, failed := latestWorkflowByPhase(creationWorkflows)

	if inProgress != nil {
		logger.V(2).Info("run configuration creation in progress")
		return []ExperimentCommand{}
	}

	newStatus := experiment.Status.DeepCopy()

	if succeeded != nil {
		logger.Info("run configuration creation succeeded")
		newStatus.SynchronizationState = pipelinesv1.Succeeded
		idResult, err := getWorkflowOutput(succeeded, ExperimentWorkflowConstants.ExperimentIdParameterName)

		if err != nil {
			logger.Error(err, "could not retrieve workflow output, failing run configuration")
			newStatus.SynchronizationState = pipelinesv1.Failed
		} else {
			newStatus.KfpId = idResult
		}
	} else {
		if failed != nil {
			logger.Info("run configuration update failed")
		} else {
			logger.Info("run configuration creation progress unknown, failing run configuration")
		}
		newStatus.SynchronizationState = pipelinesv1.Failed
	}

	return []ExperimentCommand{
		SetExperimentStatus{
			Status: *newStatus,
		},
		DeleteExperimentWorkflows{
			Workflows: creationWorkflows,
		},
	}
}
