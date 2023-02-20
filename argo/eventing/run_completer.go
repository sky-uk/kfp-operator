package eventing

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func completionStateForRunCompletionStatus(rcs RunCompletionStatus) *pipelinesv1.CompletionState {
	switch rcs {
	case RunCompletionStatuses.Succeeded:
		return &pipelinesv1.CompletionStates.Succeeded
	case RunCompletionStatuses.Failed:
		return &pipelinesv1.CompletionStates.Failed
	default:
		return nil
	}
}

type RunCompleter struct {
	K8sClient client.Client
}

func (c *RunCompleter) CompleteRun(ctx context.Context, runCompletionEvent RunCompletionEvent) error {
	run := pipelinesv1.Run{}

	err := c.K8sClient.Get(ctx, types.NamespacedName{Namespace: runCompletionEvent.Run.Namespace, Name: runCompletionEvent.Run.Name}, &run)
	if err != nil {
		return err
	}

	if completionState := completionStateForRunCompletionStatus(runCompletionEvent.Status); completionState != nil {
		run.Status.CompletionState = *completionState
		return c.K8sClient.Status().Update(ctx, &run)
	}

	return nil
}
