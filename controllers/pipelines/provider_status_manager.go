package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ProviderStatusManager interface {
	UpdateProviderStatus(ctx context.Context, provider *pipelinesv1.Provider, state apis.SynchronizationState, message string) error
}

type StatusManager struct {
	client *controllers.OptInClient
}

func (sm StatusManager) UpdateProviderStatus(ctx context.Context, provider *pipelinesv1.Provider, state apis.SynchronizationState, message string) error {
	logger := log.FromContext(ctx)

	if state == apis.Succeeded {
		provider.Status.ObservedGeneration = provider.Generation
	}
	provider.Status.SynchronizationState = state
	provider.StatusWithCondition(message)

	if err := sm.client.Status().Update(ctx, provider); err != nil {
		logger.Error(err, "Failed to update provider status", "provider", provider.GetNamespacedName(), "status", state)
		return err
	}
	return nil
}
