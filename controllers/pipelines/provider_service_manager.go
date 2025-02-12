package pipelines

import (
	"context"
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/provider/predicates"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"slices"
)

type ServiceResourceManager interface {
	Create(ctx context.Context, new *corev1.Service, provider *pipelinesv1.Provider) error
	Delete(ctx context.Context, old *corev1.Service) error
	Get(ctx context.Context, owner *pipelinesv1.Provider) (*corev1.Service, error)
	Equal(a, b *corev1.Service) bool
	Construct(provider *pipelinesv1.Provider) *corev1.Service
}

type ServiceManager struct {
	client *controllers.OptInClient
	scheme *runtime.Scheme
	config *config.KfpControllerConfigSpec
}

func (sm ServiceManager) Create(ctx context.Context, new *corev1.Service, provider *pipelinesv1.Provider) error {
	logger := log.FromContext(ctx)

	if err := ctrl.SetControllerReference(provider, new, sm.scheme); err != nil {
		logger.Error(err, "unable to set controller reference on service")
		return err
	}

	if err := sm.client.Create(ctx, new); err != nil {
		logger.Error(err, "unable to create provider service")
		return err
	}
	return nil
}

func (sm ServiceManager) Delete(ctx context.Context, old *corev1.Service) error {
	logger := log.FromContext(ctx)

	if err := sm.client.Delete(ctx, old); err != nil {
		logger.Error(err, "unable to delete existing provider service")
		return err
	}
	return nil
}

func (sm ServiceManager) Get(ctx context.Context, owner *pipelinesv1.Provider) (*corev1.Service, error) {
	sl := &corev1.ServiceList{}
	err := sm.client.NonCached.List(ctx, sl, &client.ListOptions{
		Namespace: owner.Namespace,
	})
	if err != nil {
		return nil, err
	}
	for _, svc := range sl.Items {
		if metav1.IsControlledBy(&svc, owner) {
			return &svc, nil
		}
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")

}

func (sm ServiceManager) Construct(provider *pipelinesv1.Provider) *corev1.Service {
	prefixedProviderName := fmt.Sprintf("provider-%s", provider.Name)

	ports := []corev1.ServicePort{{
		Name:       "http",
		Port:       int32(sm.config.DefaultProviderValues.ServicePort),
		TargetPort: intstr.FromInt(sm.config.DefaultProviderValues.ServicePort),
		Protocol:   corev1.ProtocolTCP,
	}}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", prefixedProviderName),
			Namespace:    provider.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{AppLabel: prefixedProviderName},
		},
	}

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	svc.Annotations[predicates.ControllerManagedKey] = "true"

	return svc
}

func (sm ServiceManager) Equal(a, b *corev1.Service) bool {
	return a.GenerateName == b.GenerateName &&
		maps.Equal(a.Spec.Selector, b.Spec.Selector) &&
		slices.Equal(a.Spec.Ports, b.Spec.Ports)
}
