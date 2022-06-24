package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

var WorkflowConstants = struct {
	OwnerKindLabelKey    string
	OwnerNameLabelKey    string
	OperationLabelKey    string
	CreateOperationLabel string
	DeleteOperationLabel string
	UpdateOperationLabel string
}{
	OwnerKindLabelKey:    pipelinesv1.GroupVersion.Group + "/owner.kind",
	OwnerNameLabelKey:    pipelinesv1.GroupVersion.Group + "/owner.name",
	OperationLabelKey:    pipelinesv1.GroupVersion.Group + "/operation",
	CreateOperationLabel: "create",
	DeleteOperationLabel: "delete",
	UpdateOperationLabel: "update",
}

type WorkflowFactory struct {
	ResourceKind string
	Config       configv1.Configuration
}

type KfpExtCommandBuilder struct {
	commandParts []string
	error        error
}

func CommonWorkflowMeta(owner types.NamespacedName, operation string, ownerKind string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-" + ownerKind + "-",
		Namespace:    owner.Namespace,
		Labels:       CommonWorkflowLabels(owner, operation, ownerKind),
	}
}

func CommonWorkflowLabels(owner types.NamespacedName, operation string, ownerKind string) map[string]string {
	return map[string]string{
		WorkflowConstants.OperationLabelKey: operation,
		WorkflowConstants.OwnerKindLabelKey: ownerKind,
		WorkflowConstants.OwnerNameLabelKey: owner.Name,
	}
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
