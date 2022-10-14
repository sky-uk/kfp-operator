package main

import (
	"context"
	"encoding/json"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/yalp/jsonpath"
	"io"
	"log"
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
	if err := RunProviderApp[KfpProviderConfig](KfpProvider{}); err != nil {
		log.Fatal(err)
	}
}

func (kfpp KfpProvider) CreatePipeline(providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, pipelineFileName string, ctx context.Context) (string, error) {
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

	return kfpp.UpdatePipeline(providerConfig, pipelineDefinition, id.(string), pipelineFileName, ctx)
}

func (kfpp KfpProvider) UpdatePipeline(providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload-version", "--pipeline-version", pipelineDefinition.Version, "--pipeline-id", id, pipelineFile)

	return id, cmd.Run()
}

func (kfpp KfpProvider) DeletePipeline(providerConfig KfpProviderConfig, id string, _ context.Context) error {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "delete", id)

	return cmd.Run()
}

func (kfpp KfpProvider) CreateRunConfiguration(providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "job", "submit",
		"--pipeline-name", runConfigurationDefinition.PipelineName,
		"--job-name", runConfigurationDefinition.Name,
		"--experiment-name", runConfigurationDefinition.ExperimentName,
		"--version-name", runConfigurationDefinition.PipelineVersion,
		"--cron-expression", runConfigurationDefinition.Schedule)

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

func (kfpp KfpProvider) UpdateRunConfiguration(providerConfig KfpProviderConfig, runConfigurationDefinition RunConfigurationDefinition, id string, ctx context.Context) (string, error) {
	if err := kfpp.DeleteRunConfiguration(providerConfig, id, ctx); err != nil {
		return id, err
	}

	return kfpp.CreateRunConfiguration(providerConfig, runConfigurationDefinition, ctx)
}

func (kfpp KfpProvider) DeleteRunConfiguration(providerConfig KfpProviderConfig, id string, _ context.Context) error {
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

func (kfpp KfpProvider) CreateExperiment(providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition, _ context.Context) (string, error) {
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

func (kfpp KfpProvider) UpdateExperiment(providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition, id string, ctx context.Context) (string, error) {
	if err := kfpp.DeleteExperiment(providerConfig, id, ctx); err != nil {
		return id, err
	}

	return kfpp.CreateExperiment(providerConfig, experimentDefinition, ctx)
}

func (kfpp KfpProvider) DeleteExperiment(providerConfig KfpProviderConfig, id string, _ context.Context) error {
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
