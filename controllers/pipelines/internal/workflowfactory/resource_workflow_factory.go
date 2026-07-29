package workflowfactory

import (
	"fmt"
	"net"
	"strconv"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	corev1 "k8s.io/api/core/v1"
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
