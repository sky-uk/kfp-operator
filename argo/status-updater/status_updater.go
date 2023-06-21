package status_updater

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (c *StatusUpdater) UpdateStatus(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
	if runCompletionEvent.RunName != nil {
		if err := c.completeRun(ctx, runCompletionEvent); err != nil {
			return err
		}
	}

	if runCompletionEvent.RunConfigurationName != nil {
		if err := c.completeRunConfiguration(ctx, runCompletionEvent); err != nil {
			return err
		}
	}

	return nil
}

func (c *StatusUpdater) completeRun(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
	if runCompletionEvent.RunName.Namespace == "" {
		return nil
	}

	run := pipelinesv1.Run{}

	if err := c.K8sClient.Get(ctx, types.NamespacedName{Namespace: runCompletionEvent.RunName.Namespace, Name: runCompletionEvent.RunName.Name}, &run); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if completionState := completionStateForRunCompletionStatus(runCompletionEvent.Status); completionState != nil {
		run.Status.CompletionState = *completionState

		if err := c.K8sClient.Status().Update(ctx, &run); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}

			return err
		}
	}

	return nil
}

func (c *StatusUpdater) completeRunConfiguration(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
	if runCompletionEvent.Status != common.RunCompletionStatuses.Succeeded || runCompletionEvent.RunConfigurationName.Namespace == "" {
		return nil
	}

	runConfiguration := pipelinesv1.RunConfiguration{}

	if err := c.K8sClient.Get(ctx, types.NamespacedName{Namespace: runCompletionEvent.RunConfigurationName.Namespace, Name: runCompletionEvent.RunConfigurationName.Name}, &runConfiguration); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	runConfiguration.Status.LatestRuns.Succeeded.ProviderId = runCompletionEvent.RunId
	runConfiguration.Status.LatestRuns.Succeeded.Artifacts = runCompletionEvent.Artifacts

	if err := c.K8sClient.Status().Update(ctx, &runConfiguration); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}
