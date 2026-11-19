package provider

import (
	"fmt"
	"slices"

	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

// kubernetesPlatform is the platform spec key Kubeflow Pipelines reads
// Kubernetes executor settings from.
const kubernetesPlatform = "kubernetes"

// KubernetesExecutorConfig returns an executor's Kubernetes settings, which are
// empty if it has none. Fields this provider does not know are left out, so
// change settings with MergeKubernetesExecutorConfig rather than the result.
func (p *PipelinePackage) KubernetesExecutorConfig(executor string) (*kubernetesplatform.KubernetesExecutorConfig, error) {
	executorConfig := &kubernetesplatform.KubernetesExecutorConfig{}
	if settings := p.kubernetesExecutorStruct(executor); settings != nil {
		if err := decodeStruct(settings, executorConfig); err != nil {
			return nil, fmt.Errorf("invalid Kubernetes settings for executor %q: %w", executor, err)
		}
	}
	return executorConfig, nil
}

// MergeKubernetesExecutorConfig adds settings to an executor's Kubernetes
// settings without overriding the pipeline's own: a field the executor already
// sets is kept, except that a list field gains the items it does not already
// have. Fields this provider does not know are kept, and settings with no
// fields set change nothing.
func (p *PipelinePackage) MergeKubernetesExecutorConfig(
	executor string,
	settings *kubernetesplatform.KubernetesExecutorConfig,
) error {
	additions, err := structFromMessage(settings)
	if err != nil {
		return fmt.Errorf("invalid Kubernetes settings: %w", err)
	}
	if len(additions.GetFields()) == 0 {
		return nil
	}
	executorConfig := p.ensureKubernetesExecutorStruct(executor)

	fields := settings.ProtoReflect().Descriptor().Fields()
	for i := range fields.Len() {
		field := fields.Get(i)
		addition, isAdded := additions.Fields[field.JSONName()]
		if !isAdded {
			continue
		}

		name, existing := kubernetesField(executorConfig, field)
		switch {
		case existing == nil:
			executorConfig.Fields[name] = addition
		case field.IsList():
			items := existing.GetListValue()
			if items == nil {
				return fmt.Errorf("%s is not a list", name)
			}
			for _, item := range addition.GetListValue().GetValues() {
				if !slices.ContainsFunc(items.Values, func(v *structpb.Value) bool { return proto.Equal(v, item) }) {
					items.Values = append(items.Values, item)
				}
			}
		}
	}
	return nil
}

// kubernetesField returns a field of an executor's Kubernetes settings and the
// name it is set under. protojson reads either the JSON or the proto name of a
// field, so both are looked up; an unset field is named by its JSON name.
func kubernetesField(executorConfig *structpb.Struct, field protoreflect.FieldDescriptor) (string, *structpb.Value) {
	for _, name := range []string{field.JSONName(), string(field.Name())} {
		if value, isSet := executorConfig.Fields[name]; isSet {
			return name, value
		}
	}
	return field.JSONName(), nil
}

// kubernetesExecutorStruct returns an executor's Kubernetes settings, or nil if
// it has none. The settings stay an untyped struct so that fields this provider
// does not know pass through unchanged.
func (p *PipelinePackage) kubernetesExecutorStruct(executor string) *structpb.Struct {
	return p.PlatformSpec.GetPlatforms()[kubernetesPlatform].GetDeploymentSpec().GetExecutors()[executor]
}

// ensureKubernetesExecutorStruct returns an executor's Kubernetes settings,
// creating the platform spec and the executor's entry if needed.
func (p *PipelinePackage) ensureKubernetesExecutorStruct(executor string) *structpb.Struct {
	if p.PlatformSpec == nil {
		p.PlatformSpec = &pipelinespec.PlatformSpec{}
	}
	if p.PlatformSpec.Platforms == nil {
		p.PlatformSpec.Platforms = map[string]*pipelinespec.SinglePlatformSpec{}
	}
	platform := p.PlatformSpec.Platforms[kubernetesPlatform]
	if platform == nil {
		platform = &pipelinespec.SinglePlatformSpec{}
		p.PlatformSpec.Platforms[kubernetesPlatform] = platform
	}
	if platform.DeploymentSpec == nil {
		platform.DeploymentSpec = &pipelinespec.PlatformDeploymentConfig{}
	}
	if platform.DeploymentSpec.Executors == nil {
		platform.DeploymentSpec.Executors = map[string]*structpb.Struct{}
	}
	executorConfig := platform.DeploymentSpec.Executors[executor]
	if executorConfig == nil {
		executorConfig = &structpb.Struct{}
		platform.DeploymentSpec.Executors[executor] = executorConfig
	}
	if executorConfig.Fields == nil {
		executorConfig.Fields = map[string]*structpb.Value{}
	}
	return executorConfig
}
