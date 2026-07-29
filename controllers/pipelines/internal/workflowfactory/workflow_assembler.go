package workflowfactory

import (
	"fmt"
	"slices"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// workflowAssembler builds the resource-agnostic parts of an argo Workflow:
// the resource-kind and provider parameters, namespace policy, provider
// service URL, workflow metadata and template names.
type workflowAssembler struct {
	config config.ConfigSpec
}

func checkResourceNamespaceAllowed(
	resourceNamespacedName types.NamespacedName,
	provider pipelineshub.Provider,
) error {
	if len(provider.Spec.AllowedNamespaces) > 0 &&
		!slices.Contains(provider.Spec.AllowedNamespaces, resourceNamespacedName.Namespace) {
		return fmt.Errorf(
			"resource %s in namespace %s is not allowed by provider %s",
			resourceNamespacedName.Name,
			resourceNamespacedName.Namespace,
			provider.Name,
		)
	}
	return nil
}

func (a workflowAssembler) commonWorkflowMeta(owner pipelineshub.Resource) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", owner.GetKind(), owner.GetName()),
		Namespace:    a.config.WorkflowNamespace,
		Labels:       workflowconstants.CommonWorkflowLabels(owner),
	}
}

func (a workflowAssembler) createTemplateName(suffix string) string {
	return fmt.Sprintf("%screate-%s", a.config.WorkflowTemplatePrefix, suffix)
}

func (a workflowAssembler) updateTemplateName(suffix string) string {
	return fmt.Sprintf("%supdate-%s", a.config.WorkflowTemplatePrefix, suffix)
}

func (a workflowAssembler) deleteTemplateName() string {
	return fmt.Sprintf("%sdelete", a.config.WorkflowTemplatePrefix)
}

func definitionParam(definitionJson string) argo.Parameter {
	return argo.Parameter{
		Name:  workflowconstants.ResourceDefinitionParameterName,
		Value: argo.AnyStringPtr(definitionJson),
	}
}

func resourceIdParam(resource pipelineshub.Resource) argo.Parameter {
	return argo.Parameter{
		Name:  workflowconstants.ResourceIdParameterName,
		Value: argo.AnyStringPtr(resource.GetStatus().Provider.Id),
	}
}

func (a workflowAssembler) providerParams(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
) ([]argo.Parameter, error) {
	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		return nil, err
	}

	return []argo.Parameter{
		{
			Name:  workflowconstants.ProviderNameParameterName,
			Value: argo.AnyStringPtr(namespacedProvider),
		},
		{
			Name: workflowconstants.ProviderServiceUrl,
			Value: argo.AnyStringPtr(
				createProviderServiceUrl(
					providerSvc,
					a.config.DefaultProviderValues.ServicePort,
				),
			),
		},
	}, nil
}

// constructWorkflow assembles an argo Workflow for a resource, applying the
// namespace policy and combining the resource-kind, provider and any extra
// parameters supplied by the caller.
func (a workflowAssembler) constructWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource pipelineshub.Resource,
	templateName string,
	extraParams []argo.Parameter,
) (*argo.Workflow, error) {
	if err := checkResourceNamespaceAllowed(resource.GetNamespacedName(), provider); err != nil {
		return nil, err
	}

	providerParams, err := a.providerParams(provider, providerSvc)
	if err != nil {
		return nil, err
	}

	params := []argo.Parameter{
		{
			Name:  workflowconstants.ResourceKindParameterName,
			Value: argo.AnyStringPtr(resource.GetKind()),
		},
	}
	params = append(params, extraParams...)
	params = append(params, providerParams...)

	return &argo.Workflow{
		ObjectMeta: *a.commonWorkflowMeta(resource),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: params,
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: templateName,
			},
		},
	}, nil
}
