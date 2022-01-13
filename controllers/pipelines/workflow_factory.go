package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type WorkflowFactory struct {
	Config configv1.Configuration
}

type KfpExtCommandBuilder struct {
	commandParts []string
	error        error
}

func escapeSingleQuotes(unescaped string) string {
	return strings.Replace(unescaped, "'", "\\'", -1)
}

func (kec *KfpExtCommandBuilder) Arg(argument string) *KfpExtCommandBuilder {
	if argument == "" {
		kec.error = fmt.Errorf("argument must not be empty")
		return kec
	}

	kec.commandParts = append(kec.commandParts, fmt.Sprintf(`'%s'`, escapeSingleQuotes(argument)))
	return kec
}

func (kec *KfpExtCommandBuilder) Param(key string, value string) *KfpExtCommandBuilder {
	if value == "" {
		kec.error = fmt.Errorf("parameter %s must not be empty", key)
		return kec
	}

	kec.commandParts = append(kec.commandParts, key)
	kec.commandParts = append(kec.commandParts, fmt.Sprintf(`'%s'`, escapeSingleQuotes(value)))
	return kec
}

func (kec *KfpExtCommandBuilder) OptParam(key string, value string) *KfpExtCommandBuilder {
	if value != "" {
		kec.commandParts = append(kec.commandParts, key)
		kec.commandParts = append(kec.commandParts, fmt.Sprintf(`'%s'`, escapeSingleQuotes(value)))
	}

	return kec
}

func (kec *KfpExtCommandBuilder) Build() (string, error) {
	return strings.Join(kec.commandParts, " "), kec.error
}

func (workflows *WorkflowFactory) KfpExt(command string) *KfpExtCommandBuilder {
	return &KfpExtCommandBuilder{
		commandParts: []string{
			"kfp-ext",
			"--endpoint",
			workflows.Config.KfpEndpoint,
			"--output",
			"json",
			command,
		},
	}
}

func (workflows *WorkflowFactory) ScriptTemplate(script string) *argo.ScriptTemplate {
	containerSpec := workflows.Config.Argo.ContainerDefaults.DeepCopy()
	containerSpec.Image = workflows.Config.Argo.KfpSdkImage
	containerSpec.Command = []string{"ash"}

	return &argo.ScriptTemplate{
		Container: *containerSpec,
		Source:    "set -e -o pipefail\n" + script,
	}
}

func (w *WorkflowFactory) Annotations(ctx context.Context, meta metav1.ObjectMeta) map[string]string {
	workflowDebugOptions := pipelinesv1.DebugOptionsFromAnnotations(ctx, meta.Annotations).WithDefaults(w.Config.Debug)
	return pipelinesv1.AnnotationsFromDebugOptions(ctx, workflowDebugOptions)
}
