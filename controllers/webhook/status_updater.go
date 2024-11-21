package webhook

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func completionStateForRunCompletionStatus(rcs common.RunCompletionStatus) *pipelinesv1.CompletionState {
	switch rcs {
	case common.RunCompletionStatuses.Succeeded:
		return &pipelinesv1.CompletionStates.Succeeded
	case common.RunCompletionStatuses.Failed:
		return &pipelinesv1.CompletionStates.Failed
	default:
		return nil
	}
}

type StatusUpdater struct {
	K8sClient client.Client
}

func NewStatusUpdater(ctx context.Context, scheme *runtime.Scheme) (StatusUpdater, error) {
	logger := log.FromContext(ctx)
	k8sConfig, err := common.K8sClientConfig()
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

func (su StatusUpdater) handle(
	ctx context.Context,
	event common.RunCompletionEvent,
) error {
	if event.RunName != nil {
		if err := su.completeRun(ctx, event); err != nil {
			return err
		}
	}
	if event.RunConfigurationName != nil {
		if err := su.completeRunConfiguration(ctx, event); err != nil {
			return err
		}
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

	run := pipelinesv1.Run{}

	if err := su.K8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: event.RunName.Namespace,
			Name:      event.RunName.Name,
		},
		&run,
	); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(
				"RunCompletionEvent's Run was not found. Skipping.",
				"RunId",
				event.RunId,
				"RunName",
				event.RunName,
				"Action",
				"Get",
			)
			return nil
		}
		return err
	}

	if completionState := completionStateForRunCompletionStatus(event.Status); completionState != nil {
		run.Status.CompletionState = *completionState
		if err := su.K8sClient.Status().Update(ctx, &run); err != nil {
			if errors.IsNotFound(err) {
				logger.Info(
					"RunCompletionEvent's Run was not found. Skipping.",
					"RunId",
					event.RunId,
					"RunName",
					event.RunName,
					"Action",
					"Update",
				)
				return nil
			}
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

	rc := pipelinesv1.RunConfiguration{}

	if err := su.K8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: event.RunConfigurationName.Namespace,
			Name:      event.RunConfigurationName.Name,
		},
		&rc,
	); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(
				"RunCompletionEvent's RunConfiguration was not found. Skipping.",
				"RunId",
				event.RunId,
				"RunConfigurationName",
				event.RunConfigurationName,
				"Action",
				"Get",
			)
			return nil
		}
		return err
	}

	rc.Status.LatestRuns.Succeeded.ProviderId = event.RunId
	rc.Status.LatestRuns.Succeeded.Artifacts = event.Artifacts

	if err := su.K8sClient.Status().Update(ctx, &rc); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(
				"RunCompletionEvent's RunConfiguration was not found. Skipping.",
				"RunId",
				event.RunId,
				"RunConfigurationName",
				event.RunConfigurationName,
				"Action",
				"Update",
			)
			return nil
		}
		return err
	}
	return nil
}
