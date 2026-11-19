package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	yamlv3 "go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"sigs.k8s.io/yaml"
)

const kubernetesPlatform = "kubernetes"

// PipelinePackage is a compiled pipeline in the form Kubeflow Pipelines
// accepts: a pipeline spec and an optional platform spec.
type PipelinePackage struct {
	PipelineSpec *pipelinespec.PipelineSpec
	PlatformSpec *pipelinespec.PlatformSpec
}

// ParsePipelinePackage reads compiler output. The first YAML document is the
// pipeline spec and an optional second document is the platform spec. TFX can
// wrap the pipeline spec in a Vertex AI PipelineJob, which is unwrapped.
//
// Documents are decoded the way the Kubeflow Pipelines API server decodes an
// upload, so the provider accepts the same input the server does.
func ParsePipelinePackage(compiled resource.CompiledPipeline, frameworkName string) (*PipelinePackage, error) {
	documents, err := decodeJSONDocuments(compiled)
	if err != nil {
		return nil, err
	}
	switch len(documents) {
	case 0:
		return nil, errors.New("compiled pipeline is empty")
	case 1, 2:
	default:
		return nil, fmt.Errorf("expected a pipeline spec and at most one platform spec, got %d documents", len(documents))
	}

	pipelineSpecJSON := documents[0]
	if strings.EqualFold(frameworkName, "tfx") {
		if pipelineSpecJSON, err = unwrapPipelineJob(pipelineSpecJSON); err != nil {
			return nil, err
		}
	}

	pipelinePackage := &PipelinePackage{PipelineSpec: &pipelinespec.PipelineSpec{}}
	if err := protojson.Unmarshal(pipelineSpecJSON, pipelinePackage.PipelineSpec); err != nil {
		return nil, fmt.Errorf("invalid pipeline spec: %w", err)
	}

	if len(documents) == 2 {
		pipelinePackage.PlatformSpec = &pipelinespec.PlatformSpec{}
		if err := protojson.Unmarshal(documents[1], pipelinePackage.PlatformSpec); err != nil {
			return nil, fmt.Errorf("invalid platform spec: %w", err)
		}
	}

	return pipelinePackage, nil
}

// Marshal writes the pipeline spec as the first YAML document and the platform
// spec, when present, as the second. This is the encoding the Kubeflow
// Pipelines API server uses for stored pipeline versions.
func (p *PipelinePackage) Marshal() (resource.CompiledPipeline, error) {
	documents := []proto.Message{p.PipelineSpec}
	if p.PlatformSpec != nil {
		documents = append(documents, p.PlatformSpec)
	}

	var encoded bytes.Buffer
	for i, document := range documents {
		asJSON, err := protojson.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("failed to encode compiled pipeline: %w", err)
		}
		asYAML, err := yaml.JSONToYAML(asJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to encode compiled pipeline: %w", err)
		}
		if i > 0 {
			encoded.WriteString("---\n")
		}
		encoded.Write(asYAML)
	}
	return resource.CompiledPipeline(encoded.Bytes()), nil
}

// DeploymentConfig decodes the pipeline spec's deployment spec, which the
// pipeline spec holds as an untyped struct. The result is for reading only:
// fields this provider does not know are dropped from it, so changes are made
// to the deployment spec struct instead.
func (p *PipelinePackage) DeploymentConfig() (*pipelinespec.PipelineDeploymentConfig, error) {
	deploymentConfig := &pipelinespec.PipelineDeploymentConfig{}
	if p.PipelineSpec.GetDeploymentSpec() == nil {
		return deploymentConfig, nil
	}
	asJSON, err := protojson.Marshal(p.PipelineSpec.GetDeploymentSpec())
	if err != nil {
		return nil, fmt.Errorf("invalid deployment spec: %w", err)
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(asJSON, deploymentConfig); err != nil {
		return nil, fmt.Errorf("invalid deployment spec: %w", err)
	}
	return deploymentConfig, nil
}

// SetAcceleratorType sets the accelerator type an executor's container
// requests, in the fields that already hold one.
func (p *PipelinePackage) SetAcceleratorType(executor string, acceleratorType string) {
	accelerator := nestedStruct(p.PipelineSpec.GetDeploymentSpec(), "executors", executor, "container", "resources", "accelerator")
	if accelerator == nil {
		return
	}
	for _, name := range []string{"resourceType", "resource_type", "type"} {
		if accelerator.Fields[name].GetStringValue() != "" {
			accelerator.Fields[name] = structpb.NewStringValue(acceleratorType)
		}
	}
}

// KubernetesExecutorConfig returns the Kubernetes platform settings for an
// executor, or nil if it has none. The settings are kept as an untyped struct
// so that fields this provider does not know are passed through unchanged.
func (p *PipelinePackage) KubernetesExecutorConfig(executor string) *structpb.Struct {
	return p.PlatformSpec.GetPlatforms()[kubernetesPlatform].GetDeploymentSpec().GetExecutors()[executor]
}

// EnsureKubernetesExecutorConfig returns the Kubernetes platform settings for
// an executor, creating the platform spec and the executor entry if needed.
func (p *PipelinePackage) EnsureKubernetesExecutorConfig(executor string) *structpb.Struct {
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

// decodeJSONDocuments splits multi-document YAML and converts each non-empty
// document to JSON.
func decodeJSONDocuments(compiled []byte) ([][]byte, error) {
	decoder := yamlv3.NewDecoder(bytes.NewReader(compiled))
	var documents [][]byte
	for {
		var document any
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode compiled pipeline: %w", err)
		}
		if document == nil {
			continue
		}
		asJSON, err := json.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("failed to decode compiled pipeline: %w", err)
		}
		documents = append(documents, asJSON)
	}
	return documents, nil
}

// unwrapPipelineJob returns the pipelineSpec of a Vertex AI PipelineJob, or the
// document unchanged if it is not wrapped.
func unwrapPipelineJob(document []byte) ([]byte, error) {
	var pipelineJob map[string]json.RawMessage
	if err := json.Unmarshal(document, &pipelineJob); err != nil {
		return nil, fmt.Errorf("invalid pipeline spec: %w", err)
	}
	if pipelineSpec, ok := pipelineJob["pipelineSpec"]; ok {
		return pipelineSpec, nil
	}
	return document, nil
}

func nestedStruct(s *structpb.Struct, keys ...string) *structpb.Struct {
	for _, key := range keys {
		s = s.GetFields()[key].GetStructValue()
	}
	return s
}
