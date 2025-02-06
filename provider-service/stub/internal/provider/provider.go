package provider

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type StubProvider struct{}

func (p *StubProvider) CreatePipeline(pd resource.PipelineDefinitionWrapper) (string, error) {
	return "", nil
}
