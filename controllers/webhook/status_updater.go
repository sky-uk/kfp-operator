package webhook

import (
	"context"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func completionStateForRunCompletionStatus(rcs common.RunCompletionStatus) *pipelineshub.CompletionState {
	switch rcs {
	case common.RunCompletionStatuses.Succeeded:
		return &pipelineshub.CompletionStates.Succeeded
	case common.RunCompletionStatuses.Failed:
		return &pipelineshub.CompletionStates.Failed
	default:
		return nil
	}
}

type StatusUpdater struct {
	K8sClient client.Client
}

func NewStatusUpdater(ctx context.Context, scheme *runtime.Scheme) (StatusUpdater, error) {
	logger := log.FromContext(ctx)
	k8sConfig, err := config.K8sClientConfig()
	if err != nil {
		logger.Error(err, "Error reading k8s client config")
		return StatusUpdater{}, err
	}

	k8sClient, err := client.New(k8sConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		logger.Error(err, "Error creating k8s client")
		return StatusUpdater{}, err
	}

	logger.Info("StatusUpdater created")
	return StatusUpdater{k8sClient}, nil
}

func (su StatusUpdater) Handle(ctx context.Context, event common.RunCompletionEvent) EventError {
	logger := log.FromContext(ctx).WithValues("RunId", event.RunId)

	runConfigurationIsSpecified := event.RunConfigurationName != nil
	runConfigurationNotFound := false
	if runConfigurationIsSpecified {
		if err := su.completeRunConfiguration(ctx, event); err != nil {
			if errors.IsNotFound(err) {
				logger.Info(
					"RunCompletionEvent's RunConfiguration was not found. Skipping.",
					"RunConfigurationName",
					event.RunConfigurationName,
					"Action",
					"Get",
				)
				runConfigurationNotFound = true
			} else {
				return &FatalError{err.Error()}
			}
		}
	}

	runIsSpecified := event.RunName != nil
	runNotFound := false
	if runIsSpecified {
		if err := su.completeRun(ctx, event); err != nil {
			if errors.IsNotFound(err) {
				logger.Info(
					"RunCompletionEvent's Run was not found. Skipping.",
					"RunName",
					event.RunName,
					"Action",
					"Get",
				)
				runNotFound = true
			} else {
				return &FatalError{err.Error()}
			}
		}
	}

	if runIsSpecified && runConfigurationIsSpecified {
		// if both specified as long as one is found it is ok
		if runNotFound && runConfigurationNotFound {
			return &MissingResourceError{"Run / RunConfiguration not found"}
		}
	} else if runConfigurationIsSpecified && runConfigurationNotFound {
		return &MissingResourceError{"RunConfiguration not found"}
	} else if runIsSpecified && runNotFound {
		return &MissingResourceError{"Run not found"}
	}

	return nil
}

func (su StatusUpdater) completeRun(ctx context.Context, event common.RunCompletionEvent) error {
	logger := log.FromContext(ctx)

	if event.RunName.Namespace == "" {
		logger.Info(
			"RunCompletionEvent's RunName namespace was empty. Skipping.",
			"RunId",
			event.RunId,
		)
		return nil
	}

	run := pipelineshub.Run{}

	if err := su.K8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: event.RunName.Namespace,
			Name:      event.RunName.Name,
		},
		&run,
	); err != nil {
		return err
	}

	if completionState := completionStateForRunCompletionStatus(event.Status); completionState != nil {
		run.Status.CompletionState = *completionState
		if err := su.K8sClient.Status().Update(ctx, &run); err != nil {
			return err
		}
	}
	return nil
}

func (su StatusUpdater) completeRunConfiguration(
	ctx context.Context,
	event common.RunCompletionEvent,
) error {
	logger := log.FromContext(ctx)

	if event.Status != common.RunCompletionStatuses.Succeeded ||
		event.RunConfigurationName.Namespace == "" {
		logger.Info(
			"RunCompletionEvent's RunConfigurationName namespace was empty. Skipping.",
			"RunId",
			event.RunId,
		)
		return nil
	}

	rc := pipelineshub.RunConfiguration{}

	if err := su.K8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: event.RunConfigurationName.Namespace,
			Name:      event.RunConfigurationName.Name,
		},
		&rc,
	); err != nil {
		return err
	}

	rc.Status.LatestRuns.Succeeded.ProviderId = event.RunId
	rc.Status.LatestRuns.Succeeded.Artifacts = event.Artifacts

	if err := su.K8sClient.Status().Update(ctx, &rc); err != nil {
		return err
	}

	return nil
}
