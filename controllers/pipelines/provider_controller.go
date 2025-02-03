package pipelines

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/runtime"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	. "github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/logkeys"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	AppLabel               = "app"
	ResourceHashAnnotation = "resource-hash"
)

type ProviderReconciler struct {
	ResourceReconciler[*pipelinesv1.Provider]
	Scheme *runtime.Scheme
}

func NewProviderReconciler(ec K8sExecutionContext, config config.KfpControllerConfigSpec) *ProviderReconciler {
	return &ProviderReconciler{
		ResourceReconciler: ResourceReconciler[*pipelinesv1.Provider]{
			EC:     ec,
			Config: config,
		},
		Scheme: ec.Scheme,
	}
}

//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipelines.kubeflow.org,resources=providers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()

	logger.V(2).Info("reconciliation started", "request", req)

	var provider = &pipelinesv1.Provider{}
	if err := r.EC.Client.NonCached.Get(ctx, req.NamespacedName, provider); err != nil {
		logger.Error(err, "unable to get provider")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	prefixedProviderName := fmt.Sprintf("provider-%s", provider.Name)

	if err := r.reconcileDeployment(ctx, provider, prefixedProviderName); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, provider, prefixedProviderName); err != nil {
		return ctrl.Result{}, err
	}

	duration := time.Since(startTime)

	if err := r.updateProviderStatus(ctx, provider); err != nil {
		logger.Error(err, "failed to update provider observedGeneration status")
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("provider %s reconciliation successful", provider.Name), logkeys.Duration, duration)

	return ctrl.Result{}, nil
}

func (r *ProviderReconciler) reconcileService(ctx context.Context, provider *pipelinesv1.Provider, prefixedProviderName string) error {
	logger := log.FromContext(ctx)
	existingSvc, err := r.getService(ctx, *provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing service")
		return err
	}

	desiredSvc, err := constructService(r.Config, provider, prefixedProviderName)
	if err != nil {
		logger.Error(err, "unable to construct provider service")
		return err
	}

	if err := ctrl.SetControllerReference(provider, desiredSvc, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference on service")
		return err
	}

	if existingSvc != nil {
		if svcIsOutOfSync(existingSvc, desiredSvc) {
			if err := r.EC.Client.Delete(ctx, existingSvc); err != nil {
				logger.Error(err, "unable to delete existing provider service")
				return err
			}
			if err := r.EC.Client.Create(ctx, desiredSvc); err != nil {
				logger.Error(err, "unable to create provider service")
				return err
			}
		}
	} else {
		if err := r.EC.Client.Create(ctx, desiredSvc); err != nil {
			logger.Error(err, "unable to create provider service")
			return err
		}
	}
	return nil
}

func (r *ProviderReconciler) reconcileDeployment(ctx context.Context, provider *pipelinesv1.Provider, prefixedProviderName string) error {
	logger := log.FromContext(ctx)

	desiredDeployment, err := constructDeployment(provider, *r.Config.DeepCopy(), prefixedProviderName)
	if err != nil {
		logger.Error(err, "unable to construct provider deployment")
		return err
	}

	if err := ctrl.SetControllerReference(provider, desiredDeployment, r.Scheme); err != nil {
		logger.Error(err, "unable to set controller reference on deployment")
		return err
	}

	logger.V(2).Info("constructed desired provider deployment", "deployment", desiredDeployment)

	existingDeployment, err := r.getDeployment(ctx, *provider)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "unable to get existing deployment")
		return err
	}

	if existingDeployment != nil {
		logger.V(2).Info("found existing provider deployment", "deployment", existingDeployment)

		if err := setDeploymentHashAnnotation(existingDeployment); err != nil {
			logger.Error(err, "unable to set resource hash annotation on existing deployment")
			return err
		}

		if deploymentIsOutOfSync(existingDeployment, desiredDeployment) {
			logger.V(2).Info("resource hash mismatch, updating deployment")
			existingDeployment = syncDeployment(existingDeployment, desiredDeployment)
			if err = r.EC.Client.Update(ctx, existingDeployment); err != nil {
				logger.Error(err, "unable to update provider deployment", "deployment", desiredDeployment)
				return err
			}
		}
	} else {
		if err = r.EC.Client.Create(ctx, desiredDeployment); err != nil {
			logger.Error(err, "unable to create provider deployment")
			return err
		}

		logger.V(2).Info("created provider deployment", "deployment", desiredDeployment)
	}
	return nil
}

func (r *ProviderReconciler) updateProviderStatus(ctx context.Context, provider *pipelinesv1.Provider) error {
	provider.Status.ObservedGeneration = provider.Generation
	if err := r.EC.Client.Status().Update(ctx, provider); err != nil {
		return err
	}
	return nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	provider := &pipelinesv1.Provider{}
	return ctrl.NewControllerManagedBy(mgr).
		For(provider).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Complete(r)
}

func constructDeployment(provider *pipelinesv1.Provider, config config.KfpControllerConfigSpec, prefixedProviderName string) (*appsv1.Deployment, error) {
	matchLabels := map[string]string{AppLabel: prefixedProviderName}
	deploymentLabels := MapConcat(config.DefaultProviderValues.Labels, matchLabels)
	replicas := int32(config.DefaultProviderValues.Replicas)

	podTemplate := config.DefaultProviderValues.PodTemplateSpec
	populatedPodTemplate, err := populateServiceContainer(config.DefaultProviderValues.ServiceContainerName, *podTemplate.DeepCopy(), provider)
	if err != nil {
		return nil, err
	}
	podTemplate = *populatedPodTemplate
	podTemplate.Spec.ServiceAccountName = provider.Spec.ServiceAccount
	podTemplate.ObjectMeta.Labels = MapConcat(podTemplate.ObjectMeta.Labels, matchLabels)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", prefixedProviderName),
			Namespace:    provider.Namespace,
			Labels:       deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Replicas: &replicas,
			Template: podTemplate,
		},
	}

	if err := setDeploymentHashAnnotation(deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func populateServiceContainer(serviceContainerName string, podTemplate v1.PodTemplateSpec, provider *pipelinesv1.Provider) (*v1.PodTemplateSpec, error) {
	if !Exists(podTemplate.Spec.Containers, func(c v1.Container) bool {
		return c.Name == serviceContainerName
	}) {

		return nil, fmt.Errorf("unable to populate service container: container with name %s not found on deployment", serviceContainerName)
	}

	podTemplate.Spec.Containers = Map(podTemplate.Spec.Containers, func(c v1.Container) v1.Container {
		if c.Name == serviceContainerName {
			c.Image = provider.Spec.ServiceImage
			c.Env = append(c.Env, v1.EnvVar{
				Name:  "PROVIDERNAME",
				Value: provider.Name,
			})
		}
		return c
	})

	return &podTemplate, nil
}

func setDeploymentHashAnnotation(deployment *appsv1.Deployment) error {
	hasher := NewObjectHasher()
	if err := hasher.WriteObject(deployment); err != nil {
		return err
	}

	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations[ResourceHashAnnotation] = fmt.Sprintf("%x", hasher.Sum())

	return nil
}

func (r *ProviderReconciler) getDeployment(ctx context.Context, provider pipelinesv1.Provider) (*appsv1.Deployment, error) {
	dl := &appsv1.DeploymentList{}
	if err := r.EC.Client.NonCached.List(ctx, dl, &client.ListOptions{
		Namespace: provider.Namespace,
	}); err != nil {
		return nil, err
	}

	for _, deploy := range dl.Items {
		if metav1.IsControlledBy(&deploy, &provider) {
			return &deploy, nil
		}
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func deploymentIsOutOfSync(existingDeployment, desiredDeployment *appsv1.Deployment) bool {
	return existingDeployment.Annotations != nil && existingDeployment.Annotations[ResourceHashAnnotation] != desiredDeployment.Annotations[ResourceHashAnnotation]
}

func syncDeployment(existingDeployment, desiredDeployment *appsv1.Deployment) *appsv1.Deployment {
	syncedDeployment := existingDeployment.DeepCopy()

	syncedDeployment.Spec = desiredDeployment.Spec
	syncedDeployment.SetLabels(desiredDeployment.Labels)
	if syncedDeployment.Annotations == nil {
		syncedDeployment.Annotations = make(map[string]string)
	}
	syncedDeployment.Annotations[ResourceHashAnnotation] = desiredDeployment.Annotations[ResourceHashAnnotation]

	return syncedDeployment
}

func (r *ProviderReconciler) getService(ctx context.Context, provider pipelinesv1.Provider) (*v1.Service, error) {
	sl := &v1.ServiceList{}
	err := r.EC.Client.NonCached.List(ctx, sl, &client.ListOptions{
		Namespace: provider.Namespace,
	})
	if err != nil {
		return nil, err
	}
	for _, svc := range sl.Items {
		if metav1.IsControlledBy(&svc, &provider) {
			return &svc, nil
		}
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func setSvcHashAnnotation(svc *v1.Service) error {
	hasher := NewObjectHasher()
	if err := hasher.WriteObject(svc); err != nil {
		return err
	}

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	svc.Annotations[ResourceHashAnnotation] = fmt.Sprintf("%x", hasher.Sum())

	return nil
}

func svcIsOutOfSync(existingSvc, desiredSvc *v1.Service) bool {
	return existingSvc.Annotations != nil && existingSvc.Annotations[ResourceHashAnnotation] != desiredSvc.Annotations[ResourceHashAnnotation]
}

func constructService(
	config config.KfpControllerConfigSpec,
	provider *pipelinesv1.Provider,
	prefixedProviderName string,
) (*v1.Service, error) {
	ports := []v1.ServicePort{{
		Name:       "http",
		Port:       int32(config.DefaultProviderValues.ServicePort),
		TargetPort: intstr.FromInt(config.DefaultProviderValues.ServicePort),
		Protocol:   v1.ProtocolTCP,
	}}
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", prefixedProviderName),
			Namespace:    provider.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports:    ports,
			Type:     v1.ServiceTypeClusterIP,
			Selector: map[string]string{AppLabel: prefixedProviderName},
		},
	}

	if err := setSvcHashAnnotation(svc); err != nil {
		return nil, err
	}

	return svc, nil
}
