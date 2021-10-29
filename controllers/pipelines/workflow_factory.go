package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowFactory struct {
	Config configv1.Configuration
}

func (workflows *WorkflowFactory) ScriptTemplate(kfpScript string) *argo.ScriptTemplate {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp-ext --endpoint %s --output json %s",
				workflows.Config.KfpEndpoint, kfpScript)

	containerSpec := workflows.Config.Argo.ContainerDefaults.DeepCopy()
	containerSpec.Image = workflows.Config.Argo.KfpSdkImage
	containerSpec.Command = []string{"ash"}

	return &argo.ScriptTemplate{
		Container: *containerSpec,
		Source:    script,
	}
}

func (w *WorkflowFactory) Annotations(ctx context.Context, meta metav1.ObjectMeta) map[string]string {
	workflowDebugOptions := pipelinesv1.DebugOptionsFromAnnotations(ctx, meta.Annotations).WithDefaults(w.Config.Debug)
	return pipelinesv1.AnnotationsFromDebugOptions(ctx, workflowDebugOptions)
}
