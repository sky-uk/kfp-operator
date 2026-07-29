package workflowfactory

import (
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

// definitionBuilder builds a resource's provider definition from the resource
// alone, requiring no provider or framework information.
type definitionBuilder[R pipelineshub.Resource, D any] interface {
	build(resource R) (D, error)
}
