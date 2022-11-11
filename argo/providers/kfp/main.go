package main

import (
	"context"
	"encoding/json"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/yalp/jsonpath"
	"io"
	"os/exec"
	"regexp"
)

const KfpResourceNotFoundCode = 5

type KfpProviderConfig struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

type KfpProvider struct {
}

func main() {
	app := ProviderApp[KfpProviderConfig]{}
	app.Run(KfpProvider{})
}

func (kfpp KfpProvider) CreatePipeline(ctx context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, pipelineFileName string) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload", "--pipeline-name", pipelineDefinition.Name, pipelineFileName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var jsonOutput interface{}
	if err = json.Unmarshal(output, &jsonOutput); err != nil {
		return "", err
	}

	id, err := jsonpath.Read(jsonOutput, `$["Pipeline Details"]["Pipeline ID"]`)
	if err != nil {
		return "", err
	}

	return kfpp.UpdatePipeline(ctx, providerConfig, pipelineDefinition, id.(string), pipelineFileName)
}

func (kfpp KfpProvider) UpdatePipeline(_ context.Context, providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload-version", "--pipeline-version", pipelineDefinition.Version, "--pipeline-id", id, pipelineFile)

	return id, cmd.Run()
}

func (kfpp KfpProvider) DeletePipeline(ctx context.Context, providerConfig KfpProviderConfig, id string) error {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "delete", id)

	return cmd.Run()
}

func (kfpp KfpProvider) CreateRunConfiguration(_ context.Context, providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition) (string, error) {
	schedule, err := ParseCron(runConfigurationDefinition.Schedule)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "job", "submit",
		"--pipeline-name", runConfigurationDefinition.PipelineName,
		"--job-name", runConfigurationDefinition.Name,
		"--experiment-name", runConfigurationDefinition.ExperimentName,
		"--version-name", runConfigurationDefinition.PipelineVersion,
		"--cron-expression", schedule.PrintGo())

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var jsonOutput interface{}
	if err = json.Unmarshal(output, &jsonOutput); err != nil {
		return "", err
	}

	id, err := jsonpath.Read(jsonOutput, `$["Job Details"]["ID"]`)
	if err != nil {
		return "", err
	}

	return id.(string), nil
}

func (kfpp KfpProvider) UpdateRunConfiguration(ctx context.Context, providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition, id string) (string, error) {
	if err := kfpp.DeleteRunConfiguration(ctx, providerConfig, id); err != nil {
		return id, err
	}

	return kfpp.CreateRunConfiguration(ctx, providerConfig, runConfigurationDefinition)
}

func (kfpp KfpProvider) DeleteRunConfiguration(_ context.Context, providerConfig KfpProviderConfig, id string) error {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "job", "delete", id)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	errOutput, err := io.ReadAll(stderr)
	if err != nil {
		return err
	}

	if cmdErr := cmd.Wait(); cmdErr != nil {
		if _, isExitError := cmdErr.(*exec.ExitError); !isExitError {
			return cmdErr
		}

		re := regexp.MustCompile(`(?m)^.*HTTP response body: ({.*})$`)
		matches := re.FindStringSubmatch(string(errOutput))

		if len(matches) < 2 {
			return cmdErr
		}

		var jsonResponse interface{}
		if err = json.Unmarshal([]byte(matches[1]), &jsonResponse); err != nil {
			return err
		}

		errorCode, err := jsonpath.Read(jsonResponse, `$["code"]`)
		if err != nil {
			return err
		}

		if int(errorCode.(float64)) != KfpResourceNotFoundCode {
			return cmdErr
		}
	}

	return nil
}

func (kfpp KfpProvider) CreateExperiment(_ context.Context, providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "experiment", "create", experimentDefinition.Name)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var jsonOutput interface{}
	if err = json.Unmarshal(output, &jsonOutput); err != nil {
		return "", err
	}

	id, err := jsonpath.Read(jsonOutput, `$["ID"]`)
	if err != nil {
		return "", err
	}

	return id.(string), nil
}

func (kfpp KfpProvider) UpdateExperiment(ctx context.Context, providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition, id string) (string, error) {
	if err := kfpp.DeleteExperiment(ctx, providerConfig, id); err != nil {
		return id, err
	}

	return kfpp.CreateExperiment(ctx, providerConfig, experimentDefinition)
}

func (kfpp KfpProvider) DeleteExperiment(_ context.Context, providerConfig KfpProviderConfig, id string) error {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "experiment", "delete", id)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if _, err = io.WriteString(stdin, "y\n"); err != nil {
		return err
	}

	return cmd.Run()
}
