package controllers

import argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

var removeString = func(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

var mapParams = func(params []argo.Parameter) map[string]string {
	m := make(map[string]string, len(params))
	for i := range params {
		m[params[i].Name] = string(*params[i].Value)
	}

	return m
}

const (
	PipelineIdKey = "pipeline-id"
)

func workflowOutput(workflow argo.Workflow, key string) string {
	globalNode, exists := workflow.Status.Nodes[workflow.Name]
	if exists && globalNode.Outputs != nil {
		return string(mapParams(globalNode.Outputs.Parameters)[key])
	}

	return ""
}

func setWorkflowOutput(workflow argo.Workflow, name string, output string) argo.Workflow {
	result := argo.AnyString(output)
	nodes := make(map[string]argo.NodeStatus)
	nodes[workflow.ObjectMeta.Name] = argo.NodeStatus{
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

func latestWorkflowsByPhase(workflows []argo.Workflow) (inProgress *argo.Workflow, succeeded *argo.Workflow, failed *argo.Workflow) {
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
