package controllers

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

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
