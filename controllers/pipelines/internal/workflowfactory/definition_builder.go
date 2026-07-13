package workflowfactory

import (
	"encoding/json"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/jsonutil"
)

// definitionBuilder builds a resource's provider definition from the resource
// alone, requiring no provider or framework information.
type definitionBuilder[R pipelineshub.Resource, D any] interface {
	build(resource R) (D, error)
}

// marshalDefinition marshals a definition to JSON and applies the given patches.
func marshalDefinition[D any](definition D, patches []pipelineshub.Patch) (string, error) {
	marshalled, err := json.Marshal(&definition)
	if err != nil {
		return "", err
	}

	return jsonutil.PatchJson(patches, marshalled)
}
