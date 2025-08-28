package workflowutil

import (
	"encoding/json"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	providers "github.com/sky-uk/kfp-operator/pkg/providers/base"
)

var mapParams = func(params []argo.Parameter) map[string]string {
	m := make(map[string]string, len(params))
	for i := range params {
		m[params[i].Name] = string(*params[i].Value)
	}

	return m
}

func GetWorkflowParameter(workflow *argo.Workflow, name string) string {
	for _, parameter := range workflow.Spec.Arguments.Parameters {
		if parameter.Name == name {
			return parameter.Value.String()
		}
	}

	return ""
}

func GetWorkflowOutput(
	workflow *argo.Workflow,
	key string,
) (providers.Output, error) {
	output := providers.Output{}

	entrypointNode, exists := workflow.Status.Nodes[workflow.Name]
	if !exists || entrypointNode.Outputs == nil {
		return output, fmt.Errorf("workflow does not have %s node", workflow.Name)
	}

	value, ok := mapParams(entrypointNode.Outputs.Parameters)[key]
	if !ok {
		return output, fmt.Errorf("workflow %s outputs does not have the key %s", workflow.Name, key)
	}

	err := json.Unmarshal([]byte(value), &output)

	return output, err
}

func SetWorkflowProvider(
	workflow *argo.Workflow,
	provider pipelineshub.Provider,
) (*argo.Workflow, error) {
	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		fmt.Println("WorkflowUtils: err: ", err)
		return nil, err
	}

	workflow.Spec.Arguments.Parameters = append(
		workflow.Spec.Arguments.Parameters,
		argo.Parameter{
			Name:  workflowconstants.ProviderNameParameterName,
			Value: argo.AnyStringPtr(namespacedProvider),
		},
	)

	return workflow, nil
}

func setWorkflowOutputs(workflow *argo.Workflow, parameters []argo.Parameter) *argo.Workflow {
	nodes := make(map[string]argo.NodeStatus)

	nodes[workflow.Name] = argo.NodeStatus{
		Outputs: &argo.Outputs{
			Parameters: parameters,
		},
	}

	workflow.Status.Nodes = nodes

	return workflow
}

func SetProviderOutput(workflow *argo.Workflow, output providers.Output) *argo.Workflow {
	return setWorkflowOutputs(
		workflow,
		[]argo.Parameter{
			{
				Name:  workflowconstants.ProviderOutputParameterName,
				Value: argo.AnyStringPtr(`{"id":"` + output.Id + `","providerError":"` + output.ProviderError + `"}`),
			},
		},
	)
}

func latestWorkflow(workflow1 *argo.Workflow, workflow2 *argo.Workflow) *argo.Workflow {
	if workflow1 == nil {
		return workflow2
	} else if workflow2 == nil || workflow2.ObjectMeta.CreationTimestamp.Before(&workflow1.ObjectMeta.CreationTimestamp) {
		return workflow1
	} else {
		return workflow2
	}
}

func LatestWorkflowByPhase(
	workflows []argo.Workflow,
) (inProgress *argo.Workflow, succeeded *argo.Workflow, failed *argo.Workflow) {

	for i := range workflows {
		workflow := workflows[i]
		switch workflow.Status.Phase {
		case argo.WorkflowFailed, argo.WorkflowError:
			failed = latestWorkflow(failed, &workflow)
		case argo.WorkflowSucceeded:
			succeeded = latestWorkflow(succeeded, &workflow)
		default:
			inProgress = latestWorkflow(inProgress, &workflow)
		}
	}

	return
}
