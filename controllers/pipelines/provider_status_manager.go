package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ProviderStatusManager struct {
	client *controllers.OptInClient
}

func (sm ProviderStatusManager) UpdateStatus(
	ctx context.Context,
	provider *pipelineshub.Provider,
	state apis.SynchronizationState,
	message string,
) error {
	logger := log.FromContext(ctx)

	if state == apis.Succeeded {
		provider.Status.ObservedGeneration = provider.Generation
	}
	provider.StatusWithCondition(state, message)

	if err := sm.client.Status().Update(ctx, provider); err != nil {
		logger.Error(err, "Failed to update provider status", "provider", provider.GetNamespacedName(), "status", state)
		return err
	}
	return nil
}
