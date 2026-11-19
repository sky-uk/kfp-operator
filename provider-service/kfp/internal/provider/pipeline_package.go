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

// PipelinePackage is a compiled pipeline in the form Kubeflow Pipelines
// accepts: a pipeline spec and an optional platform spec.
type PipelinePackage struct {
	PipelineSpec *pipelinespec.PipelineSpec
	PlatformSpec *pipelinespec.PlatformSpec
}

// ParsePipelinePackage reads compiler output. The first YAML document is the
// pipeline spec and an optional second document is the platform spec. TFX
// wraps its pipeline spec in a Vertex AI PipelineJob, which ParsePipelinePackage
// unwraps.
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

	// Reject unknown fields, as the Kubeflow Pipelines API server does on
	// upload, so that Marshal cannot drop one without an error.
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

	// The Kubeflow Pipelines API server stores pipeline versions with protojson
	// then JSONToYAML. This path avoids yaml.v3's encoder, which writes the wrong
	// indentation indicator for block scalars that start with whitespace.
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

// decodeJSONDocuments splits multi-document YAML and converts each non-empty
// document to JSON for protojson. The Kubeflow Pipelines API server reads an
// upload the same way, with yaml.v3 then encoding/json.
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
// document unchanged if it has no PipelineJob wrapper.
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

// decodeStruct reads an untyped struct into message, dropping fields the
// message does not have.
func decodeStruct(s *structpb.Struct, message proto.Message) error {
	asJSON, err := protojson.Marshal(s)
	if err != nil {
		return err
	}
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(asJSON, message)
}

// structFromMessage converts a message to the untyped struct the pipeline and
// platform specs hold their settings as.
func structFromMessage(message proto.Message) (*structpb.Struct, error) {
	asJSON, err := protojson.Marshal(message)
	if err != nil {
		return nil, err
	}
	s := &structpb.Struct{}
	if err := protojson.Unmarshal(asJSON, s); err != nil {
		return nil, err
	}
	return s, nil
}
