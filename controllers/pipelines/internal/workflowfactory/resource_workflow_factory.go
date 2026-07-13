package workflowfactory

import (
	"fmt"
	"net"
	"slices"
	"strconv"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type WorkflowFactory[R pipelineshub.Resource] interface {
	ConstructCreationWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)

	ConstructUpdateWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)

	ConstructDeletionWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)
}

func createProviderServiceUrl(svc corev1.Service, port int) string {
	return net.JoinHostPort(fmt.Sprintf("%s.%s", svc.Name, svc.Namespace), strconv.Itoa(port))
}

const (
	CompiledSuffix = "compiled"
	SimpleSuffix   = "simple"
)

func checkResourceNamespaceAllowed(
	resourceNamespacedName types.NamespacedName,
	provider pipelineshub.Provider,
) error {
	if len(provider.Spec.AllowedNamespaces) > 0 && !slices.Contains(provider.Spec.AllowedNamespaces, resourceNamespacedName.Namespace) {
		return fmt.Errorf("resource %s in namespace %s is not allowed by provider %s", resourceNamespacedName.Name, resourceNamespacedName.Namespace, provider.Name)
	}
	return nil
}
