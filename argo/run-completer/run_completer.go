package run_completer

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

type RunCompleter struct {
	K8sClient client.Client
}

func (c *RunCompleter) CompleteRun(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
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
