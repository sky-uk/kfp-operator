package workflowfactory

import (
	"fmt"
	"net"
	"slices"
	"strconv"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BaseWorkflowBuilder handles the common workflow construction logic
type BaseWorkflowBuilder struct {
	config config.KfpControllerConfigSpec
}

// NewBaseWorkflowBuilder creates a new WorkflowBuilder
func NewBaseWorkflowBuilder(config config.KfpControllerConfigSpec) *BaseWorkflowBuilder {
	return &BaseWorkflowBuilder{config: config}
}

// BuildWorkflow constructs an Argo workflow with the provided parameters
func (bwb *BaseWorkflowBuilder) BuildWorkflow(
	resource pipelineshub.Resource,
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	templateName string,
	baseParams []argo.Parameter,
	additionalParams []argo.Parameter,
) (*argo.Workflow, error) {
	if err := bwb.checkResourceNamespaceAllowed(resource, provider); err != nil {
		return nil, err
	}

	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider namespaced name: %w", err)
	}

	// Build common parameters
	commonParams := []argo.Parameter{
		{
			Name:  workflowconstants.ResourceKindParameterName,
			Value: argo.AnyStringPtr(resource.GetKind()),
		},
		{
			Name:  workflowconstants.ProviderNameParameterName,
			Value: argo.AnyStringPtr(namespacedProvider),
		},
		{
			Name: workflowconstants.ProviderServiceUrl,
			Value: argo.AnyStringPtr(bwb.createProviderServiceUrl(
				providerSvc,
				bwb.config.DefaultProviderValues.ServicePort,
			)),
		},
	}

	// Combine all parameters
	allParams := append(commonParams, baseParams...)
	allParams = append(allParams, additionalParams...)

	return &argo.Workflow{
		ObjectMeta: *bwb.commonWorkflowMeta(resource),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: allParams,
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: templateName,
			},
		},
	}, nil
}

// BuildCreationParams builds parameters specific to creation workflows
func (bwb *BaseWorkflowBuilder) BuildCreationParams(resourceDefinition string) []argo.Parameter {
	return []argo.Parameter{
		{
			Name:  workflowconstants.ResourceDefinitionParameterName,
			Value: argo.AnyStringPtr(resourceDefinition),
		},
	}
}

// BuildUpdateParams builds parameters specific to update workflows
func (bwb *BaseWorkflowBuilder) BuildUpdateParams(resourceDefinition string, resourceId string) []argo.Parameter {
	return []argo.Parameter{
		{
			Name:  workflowconstants.ResourceDefinitionParameterName,
			Value: argo.AnyStringPtr(resourceDefinition),
		},
		{
			Name:  workflowconstants.ResourceIdParameterName,
			Value: argo.AnyStringPtr(resourceId),
		},
	}
}

// BuildDeletionParams builds parameters specific to deletion workflows
func (bwb *BaseWorkflowBuilder) BuildDeletionParams(resourceId string) []argo.Parameter {
	return []argo.Parameter{
		{
			Name:  workflowconstants.ResourceIdParameterName,
			Value: argo.AnyStringPtr(resourceId),
		},
	}
}

// commonWorkflowMeta creates common metadata for workflows
func (bwb *BaseWorkflowBuilder) commonWorkflowMeta(owner pipelineshub.Resource) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", owner.GetKind(), owner.GetName()),
		Namespace:    bwb.config.WorkflowNamespace,
		Labels:       workflowconstants.CommonWorkflowLabels(owner),
	}
}

// createProviderServiceUrl creates the provider service URL
func (bwb *BaseWorkflowBuilder) createProviderServiceUrl(svc corev1.Service, port int) string {
	return net.JoinHostPort(fmt.Sprintf("%s.%s", svc.Name, svc.Namespace), strconv.Itoa(port))
}

// checkResourceNamespaceAllowed validates that the resource namespace is allowed by the provider
func (bwb *BaseWorkflowBuilder) checkResourceNamespaceAllowed(
	resource pipelineshub.Resource,
	provider pipelineshub.Provider,
) error {
	resourceNamespacedName := resource.GetNamespacedName()
	if len(provider.Spec.AllowedNamespaces) > 0 {
		if slices.Contains(provider.Spec.AllowedNamespaces, resourceNamespacedName.Namespace) {
			return nil
		}
		return fmt.Errorf("resource %s in namespace %s is not allowed by provider %s",
			resourceNamespacedName.Name, resourceNamespacedName.Namespace, provider.Name)
	}
	return nil
}
