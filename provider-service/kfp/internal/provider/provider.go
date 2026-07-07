package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type KfpProvider struct {
	config                *config.Config
	pipelineUploadService PipelineUploadService
	pipelineService       PipelineService
	runService            RunService
	recurringRunService   RecurringRunService
	experimentService     ExperimentService
	labelService          LabelService
}

func NewKfpProvider(config *config.Config) (*KfpProvider, error) {
	pipelineUploadService, err := NewPipelineUploadService(
		config.Parameters.RestKfpApiUrl,
	)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		config.Parameters.GrpcKfpApiAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	pipelineService, err := NewPipelineService(conn)
	if err != nil {
		return nil, err
	}

	labelGenerator := label.DefaultLabelGen{
		ProviderName: config.ProviderName,
	}

	runService, err := NewRunService(conn, labelGenerator, config.PipelineRootStorage)
	if err != nil {
		return nil, err
	}

	recurringRunService, err := NewRecurringRunService(conn, labelGenerator, config.PipelineRootStorage)
	if err != nil {
		return nil, err
	}

	experimentService, err := NewExperimentService(conn)
	if err != nil {
		return nil, err
	}

	labelService, err := NewDefaultLabelService()
	if err != nil {
		return nil, err
	}

	return &KfpProvider{
		config:                config,
		pipelineUploadService: pipelineUploadService,
		pipelineService:       pipelineService,
		runService:            runService,
		recurringRunService:   recurringRunService,
		experimentService:     experimentService,
		labelService:          labelService,
	}, nil
}

var _ resource.Provider = &KfpProvider{}

// extractPipelineSpec extracts the bare KFP IR PipelineSpec from a compiled pipeline.
// TFX compiles pipelines into a Vertex PipelineJob wrapper containing a nested
// pipelineSpec. For TFX pipelines, the nested pipelineSpec is extracted.
// For non-TFX pipelines (e.g. kfp-sdk), the compiled pipeline is returned unchanged.
func extractPipelineSpec(compiledPipeline json.RawMessage, frameworkName string) (json.RawMessage, error) {
	if strings.ToLower(frameworkName) != "tfx" {
		return compiledPipeline, nil
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(compiledPipeline, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TFX pipeline spec: %w", err)
	}

	if pipelineSpec, ok := wrapper["pipelineSpec"]; ok {
		return pipelineSpec, nil
	}

	// No wrapper detected, return as-is (already a bare pipelineSpec)
	return compiledPipeline, nil
}

func (p *KfpProvider) CreatePipeline(
	ctx context.Context,
	pdw resource.PipelineDefinitionWrapper,
) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(pdw.PipelineDefinition.Name)
	if err != nil {
		return "", fmt.Errorf("failed to fetch resource name %v", err)
	}

	pdw.CompiledPipeline, err = extractPipelineSpec(pdw.CompiledPipeline, pdw.PipelineDefinition.Framework.Name)
	if err != nil {
		return "", fmt.Errorf("failed to unwrap TFX pipeline spec %v", err)
	}

	pdw.CompiledPipeline, err = p.labelService.InsertLabelsIntoParameters(pdw.CompiledPipeline, label.LabelKeys)
	if err != nil {
		return "", fmt.Errorf("failed to insert labels into parameters %v", err)
	}

	log.FromContext(ctx).Info("Creating pipeline", "name", pipelineName, "spec", string(pdw.CompiledPipeline))

	pipelineId, err := p.pipelineUploadService.UploadPipeline(
		ctx,
		pdw.CompiledPipeline,
		pipelineName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload pipeline %v", err)
	}
	return p.UpdatePipeline(ctx, pdw, pipelineId)
}

func (p *KfpProvider) UpdatePipeline(
	ctx context.Context,
	pdw resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if err := p.pipelineService.DeletePipelineVersions(ctx, id); err != nil {
		return "", fmt.Errorf("failed to delete pipeline versions %v", err)
	}

	var err error
	pdw.CompiledPipeline, err = extractPipelineSpec(pdw.CompiledPipeline, pdw.PipelineDefinition.Framework.Name)
	if err != nil {
		return "", fmt.Errorf("failed to unwrap TFX pipeline spec %v", err)
	}

	pdw.CompiledPipeline, err = p.labelService.InsertLabelsIntoParameters(pdw.CompiledPipeline, label.LabelKeys)
	if err != nil {
		return "", fmt.Errorf("failed to insert labels into parameters %v", err)
	}

	log.FromContext(ctx).Info("Updating pipeline", "name", pdw.PipelineDefinition.Name, "spec", string(pdw.CompiledPipeline))

	if err := p.pipelineUploadService.UploadPipelineVersion(
		ctx,
		id,
		pdw.CompiledPipeline,
		pdw.PipelineDefinition.Version,
	); err != nil {
		return "", fmt.Errorf("failed to upload pipeline version %v", err)
	}

	return id, nil
}

func (p *KfpProvider) DeletePipeline(
	ctx context.Context,
	id string,
) error {
	if err := p.pipelineService.DeletePipelineVersions(ctx, id); err != nil {
		return err
	}

	return p.pipelineService.DeletePipeline(ctx, id)
}

func (p *KfpProvider) resolveExperimentName(
	name common.NamespacedName,
) common.NamespacedName {
	if name.Name == "" {
		return common.NamespacedName{Name: p.config.Parameters.DefaultExperiment}
	}
	return name
}

func (p *KfpProvider) CreateRun(
	ctx context.Context,
	rd base.RunDefinition,
) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := p.pipelineService.PipelineIdForDisplayName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := p.pipelineService.PipelineVersionIdForDisplayName(
		ctx,
		rd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentName := p.resolveExperimentName(rd.ExperimentName)
	experimentId, err := p.experimentService.ExperimentIdByDisplayName(ctx, experimentName)
	if err != nil {
		return "", err
	}

	runId, err := p.runService.CreateRun(
		ctx,
		rd,
		pipelineId,
		pipelineVersionId,
		experimentId,
	)
	if err != nil {
		return "", err
	}

	return runId, nil
}

func (*KfpProvider) DeleteRun(
	_ context.Context,
	_ string,
) error {
	return nil
}

func (p *KfpProvider) CreateRunSchedule(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rsd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := p.pipelineService.PipelineIdForDisplayName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := p.pipelineService.PipelineVersionIdForDisplayName(
		ctx,
		rsd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentName := p.resolveExperimentName(rsd.ExperimentName)
	experimentId, err := p.experimentService.ExperimentIdByDisplayName(ctx, experimentName)
	if err != nil {
		return "", err
	}

	recurringRunId, err := p.recurringRunService.CreateRecurringRun(
		ctx,
		rsd,
		pipelineId,
		pipelineVersionId,
		experimentId,
	)
	if err != nil {
		return "", err
	}

	return recurringRunId, nil
}

func (p *KfpProvider) UpdateRunSchedule(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
	id string,
) (string, error) {
	if err := p.DeleteRunSchedule(ctx, id); err != nil {
		return id, err
	}

	return p.CreateRunSchedule(ctx, rsd)
}

func (p *KfpProvider) DeleteRunSchedule(
	ctx context.Context,
	id string,
) error {
	return p.recurringRunService.DeleteRecurringRun(ctx, id)
}

func (p *KfpProvider) CreateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
) (string, error) {
	expId, err := p.experimentService.CreateExperiment(
		ctx,
		ed.Name,
		ed.Description,
	)
	if err != nil {
		return "", err
	}

	return expId, nil
}

func (p *KfpProvider) UpdateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
	id string,
) (string, error) {
	if err := p.DeleteExperiment(ctx, id); err != nil {
		return id, err
	}

	return p.CreateExperiment(ctx, ed)
}

func (p *KfpProvider) DeleteExperiment(
	ctx context.Context,
	id string,
) error {
	return p.experimentService.DeleteExperiment(ctx, id)
}
