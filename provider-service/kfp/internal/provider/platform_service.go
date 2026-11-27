package provider

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/protobuf/proto"
)

// PlatformService adds the Provider's platform settings to a compiled pipeline.
type PlatformService interface {
	Apply(resource.CompiledPipeline) (resource.CompiledPipeline, error)
}

// DefaultPlatformService is the default PlatformService implementation. It
// places accelerators and merges Kubernetes executor settings into the pipeline.
type DefaultPlatformService struct {
	platform config.Platform
}

var _ PlatformService = DefaultPlatformService{}

// NewDefaultPlatformService returns a DefaultPlatformService for the given platform settings.
func NewDefaultPlatformService(platform config.Platform) DefaultPlatformService {
	return DefaultPlatformService{platform: platform}
}

// Apply places accelerators, then merges the platform's Kubernetes executor
// settings into every container executor. Without accelerators or Kubernetes
// settings, the pipeline is returned unchanged and unparsed.
func (ps DefaultPlatformService) Apply(
	compiledPipeline resource.CompiledPipeline,
) (resource.CompiledPipeline, error) {
	if len(ps.platform.Accelerators) == 0 && proto.Size(ps.platform.Kubernetes) == 0 {
		return compiledPipeline, nil
	}

	pipelinePackage, err := ParsePipelinePackage(compiledPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compiled pipeline: %w", err)
	}

	// Accelerators go first, so their node selectors take precedence over the platform's.
	if err := placeAccelerators(ps.platform.Accelerators, pipelinePackage); err != nil {
		return nil, fmt.Errorf("failed to place accelerators: %w", err)
	}

	if err := pipelinePackage.MergeKubernetesExecutorConfigIntoContainers(ps.platform.Kubernetes); err != nil {
		return nil, fmt.Errorf("failed to apply platform settings: %w", err)
	}

	return pipelinePackage.Marshal()
}
