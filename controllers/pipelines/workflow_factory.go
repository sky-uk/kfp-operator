package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	apiv1 "k8s.io/api/core/v1"
)

type WorkflowFactory struct {
	Config configv1.Configuration
}

func (workflows *WorkflowFactory) ScriptTemplate(kfpScript string) *argo.ScriptTemplate {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp-ext --endpoint %s --output json %s",
				workflows.Config.KfpEndpoint, kfpScript)

	return &argo.ScriptTemplate{
		Container: apiv1.Container{
			Image:           workflows.Config.KfpSdkImage,
			ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
			Command:         []string{"ash"},
		},
		Source: script,
	}
}
