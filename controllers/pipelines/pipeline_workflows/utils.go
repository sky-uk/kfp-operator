package pipeline_workflows

import (
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

var mapParams = func(params []argo.Parameter) map[string]string {
	m := make(map[string]string, len(params))
	for i := range params {
		m[params[i].Name] = string(*params[i].Value)
	}

	return m
}

func GetWorkflowOutput(workflow *argo.Workflow, key string) (string, error) {
	entrypointNode, exists := workflow.Status.Nodes[workflow.Name]
	if exists && entrypointNode.Outputs != nil {
		return string(mapParams(entrypointNode.Outputs.Parameters)[key]), nil
	}

	return "", fmt.Errorf("workflow does not have %s node", workflow.Name)
}

func SetWorkflowOutput(workflow *argo.Workflow, name string, output string) *argo.Workflow {
	result := argo.AnyString(output)
	nodes := make(map[string]argo.NodeStatus)
	nodes[workflow.Name] = argo.NodeStatus{
		Outputs: &argo.Outputs{
			Parameters: []argo.Parameter{
				{
					Name:  name,
					Value: &result,
				},
			},
		},
	}

	workflow.Status.Nodes = nodes

	return workflow
}

func GetLatestWorkflowsByPhase(workflows []argo.Workflow) (inProgress *argo.Workflow, succeeded *argo.Workflow, failed *argo.Workflow) {
	for i := range workflows {
		workflow := workflows[i]
		switch workflow.Status.Phase {
		case argo.WorkflowFailed, argo.WorkflowError:
			failed = &workflow
		case argo.WorkflowSucceeded:
			succeeded = &workflow
		default:
			inProgress = &workflow
		}
	}

	return
}
