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

// kubernetesPlatform is the platform spec key Kubeflow Pipelines reads
// Kubernetes executor settings from.
const kubernetesPlatform = "kubernetes"

// PipelinePackage is a compiled pipeline in the form Kubeflow Pipelines
// accepts: a pipeline spec and an optional platform spec.
type PipelinePackage struct {
	PipelineSpec *pipelinespec.PipelineSpec
	PlatformSpec *pipelinespec.PlatformSpec
}

// ParsePipelinePackage reads compiler output. The first YAML document is the
// pipeline spec and an optional second document is the platform spec. TFX
// wraps its pipeline spec in a Vertex AI PipelineJob, which is unwrapped.
func ParsePipelinePackage(compiled resource.CompiledPipeline, frameworkName string) (*PipelinePackage, error) {
	documents, err := decodeJSONDocuments(compiled)
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, errors.New("compiled pipeline is empty")
	}
	if len(documents) > 2 {
		return nil, fmt.Errorf(
			"expected a pipeline spec and at most one platform spec, got %d documents",
			len(documents),
		)
	}

	pipelineSpecJSON := documents[0]
	if strings.EqualFold(frameworkName, "tfx") {
		if pipelineSpecJSON, err = unwrapPipelineJob(pipelineSpecJSON); err != nil {
			return nil, err
		}
	}

	// Unknown fields are rejected, as the Kubeflow Pipelines API server rejects
	// them on upload, so nothing is silently dropped when the package is written.
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
// spec, when there is one, as the second.
func (p *PipelinePackage) Marshal() (resource.CompiledPipeline, error) {
	documents := []proto.Message{p.PipelineSpec}
	if p.PlatformSpec != nil {
		documents = append(documents, p.PlatformSpec)
	}

	// protojson then JSONToYAML is how the Kubeflow Pipelines API server stores
	// pipeline versions. It also avoids yaml.v3's encoder, which gives block
	// scalars that start with whitespace the wrong indentation indicator.
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

// deploymentConfig decodes the pipeline spec's deployment spec, which the
// pipeline spec holds as an untyped struct. The result is only for reading:
// fields this provider does not know are dropped from it, so changes are made
// to the deployment spec itself.
func (p *PipelinePackage) deploymentConfig() (*pipelinespec.PipelineDeploymentConfig, error) {
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

// setAcceleratorType replaces the accelerator type an executor's container
// requests, in each field that holds one.
func (p *PipelinePackage) setAcceleratorType(executor string, acceleratorType string) {
	accelerator := nestedStruct(p.PipelineSpec.GetDeploymentSpec(), "executors", executor, "container", "resources", "accelerator")
	if accelerator == nil {
		return
	}
	// KFP SDK writes the deprecated type alongside resourceType, and protojson
	// reads either the JSON or the proto name of a field.
	for _, name := range []string{"resourceType", "resource_type", "type"} {
		if accelerator.Fields[name].GetStringValue() != "" {
			accelerator.Fields[name] = structpb.NewStringValue(acceleratorType)
		}
	}
}

// kubernetesExecutorConfig returns the Kubernetes settings for an executor, or
// nil if it has none. The settings stay an untyped struct so that fields this
// provider does not know pass through unchanged.
func (p *PipelinePackage) kubernetesExecutorConfig(executor string) *structpb.Struct {
	return p.PlatformSpec.GetPlatforms()[kubernetesPlatform].GetDeploymentSpec().GetExecutors()[executor]
}

// ensureKubernetesExecutorConfig returns the Kubernetes settings for an
// executor, creating the platform spec and the executor's entry if needed.
func (p *PipelinePackage) ensureKubernetesExecutorConfig(executor string) *structpb.Struct {
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
// document to JSON for protojson. Decoding with yaml.v3 then encoding/json
// matches how the Kubeflow Pipelines API server reads an upload.
func decodeJSONDocuments(compiled []byte) ([][]byte, error) {
	decoder := yamlv3.NewDecoder(bytes.NewReader(compiled))
	var documents [][]byte
	for {
		var document any
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			return documents, nil
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
}

// unwrapPipelineJob returns the pipelineSpec of a Vertex AI PipelineJob, or the
// document unchanged if it is not wrapped.
func unwrapPipelineJob(document []byte) ([]byte, error) {
	var pipelineJob map[string]json.RawMessage
	if err := json.Unmarshal(document, &pipelineJob); err != nil {
		return nil, fmt.Errorf("invalid pipeline spec: %w", err)
	}
	pipelineSpec, isWrapped := pipelineJob["pipelineSpec"]
	if !isWrapped {
		return document, nil
	}
	return pipelineSpec, nil
}

// nestedStruct follows keys through nested structs, returning nil if any key is
// missing or does not hold a struct.
func nestedStruct(s *structpb.Struct, keys ...string) *structpb.Struct {
	for _, key := range keys {
		s = s.GetFields()[key].GetStructValue()
	}
	return s
}
