package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/file"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type VAIProvider struct {
	ctx            context.Context
	config         VAIProviderConfig
	fileHandler    file.FileHandler
	pipelineClient aiplatform.PipelineClient
}

func NewProvider(
	ctx context.Context,
	config VAIProviderConfig,
) (*VAIProvider, error) {
	fh, err := file.NewGcsFileHandler(ctx, config.Parameters.GcsEndpoint)
	if err != nil {
		return nil, err
	}

	pc, err := aiplatform.NewPipelineClient(
		ctx,
		option.WithEndpoint(config.VaiEndpoint()),
	)

	return &VAIProvider{
		ctx:            ctx,
		config:         config,
		fileHandler:    &fh,
		pipelineClient: *pc,
	}, nil
}

func (vaip *VAIProvider) CreatePipeline(
	pd resource.PipelineDefinition,
) (string, error) {
	pipelineId, err := vaip.UpdatePipeline(pd, "")
	if err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) UpdatePipeline(
	pd resource.PipelineDefinition,
	id string,
) (string, error) {
	pipelineId, err := pd.Name.String()
	if err != nil {
		return "", err
	}

	storageObject, err := vaip.config.pipelineStorageObject(pd.Name, pd.Version)
	if err != nil {
		return pipelineId, err
	}

	if err = vaip.fileHandler.Write(
		pd.Manifest,
		vaip.config.Parameters.PipelineBucket,
		storageObject,
	); err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) DeletePipeline(id string) error {
	if err := vaip.fileHandler.Delete(
		id,
		vaip.config.Parameters.PipelineBucket,
	); err != nil {
		return err
	}
	return nil
}

func (vaip *VAIProvider) CreateRun(rd resource.RunDefinition) (string, error) {
	params := make(map[string]*aiplatformpb.Value, len(rd.RuntimeParameters))
	for name, value := range rd.RuntimeParameters {
		params[name] = &aiplatformpb.Value{
			Value: &aiplatformpb.Value_StringValue{
				StringValue: value,
			},
		}
	}

	templateUri, err := vaip.config.pipelineUri(
		rd.PipelineName,
		rd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	pipelinePath, err := vaip.config.pipelineStorageObject(
		rd.PipelineName,
		rd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	// skip extractBucketAndObjectFromGCSPath
	raw, err := vaip.fileHandler.Read(
		vaip.config.Parameters.PipelineBucket,
		pipelinePath,
	)
	pipelineSpec, err := structpb.NewStruct(
		raw["pipelineSpec"].(map[string]interface{}),
	)
	if err != nil {
		return "", err
	}

	job := &aiplatformpb.PipelineJob{
		DisplayName:  raw["displayName"].(string),
		PipelineSpec: pipelineSpec,
		Labels:       runLabelsFromRunDefinition(rd),
		RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
			Parameters:         params,
			GcsOutputDirectory: raw["runtimeConfig"].(map[string]any)["gcsOutputDirectory"].(string),
		},
		ServiceAccount: vaip.config.Parameters.VaiJobServiceAccount,
		TemplateUri:    templateUri,
	}

	labels := raw["labels"].(map[string]any)
	if job.Labels == nil {
		// redundant?
		job.Labels = map[string]string{}
	}
	for k, v := range labels {
		job.Labels[k] = v.(string)
	}

	runId := fmt.Sprintf("%s-%s", rd.Name.Namespace, rd.Name.Name)

	req := &aiplatformpb.CreatePipelineJobRequest{
		Parent:        vaip.config.parent(),
		PipelineJobId: fmt.Sprintf("%s-%s", runId, rd.Version),
		PipelineJob:   job,
	}

	_, err = vaip.pipelineClient.CreatePipelineJob(vaip.ctx, req)
	if err != nil {
		return "", err
	}

	return runId, nil
}

func (vaip *VAIProvider) DeleteRun(_ string) error {
	return nil
}

func (vaip *VAIProvider) CreateRunSchedule(
	rsd resource.RunScheduleDefinition,
) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	id string,
) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteRunSchedule(id string) error {
	return nil
}

func (vaip *VAIProvider) CreateExperiment(
	_ resource.ExperimentDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip *VAIProvider) UpdateExperiment(
	_ resource.ExperimentDefinition,
	_ string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip *VAIProvider) DeleteExperiment(_ string) error {
	return errors.New("not implemented")
}

func runLabelsFromPipeline(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) map[string]string {
	return map[string]string{
		labels.PipelineName:      pipelineName.Name,
		labels.PipelineNamespace: pipelineName.Namespace,
		labels.PipelineVersion:   strings.ReplaceAll(pipelineVersion, ".", "-"),
	}
}

func runLabelsFromRunDefinition(
	rd resource.RunDefinition,
) map[string]string {
	runLabels := runLabelsFromPipeline(
		rd.PipelineName,
		rd.PipelineVersion,
	)

	if !rd.RunConfigurationName.Empty() {
		runLabels[labels.RunConfigurationName] =
			rd.RunConfigurationName.Name
		runLabels[labels.RunConfigurationNamespace] =
			rd.RunConfigurationName.Namespace
	}

	if !rd.Name.Empty() {
		runLabels[labels.RunName] = rd.Name.Name
		runLabels[labels.RunNamespace] = rd.Name.Namespace
	}

	return runLabels
}
