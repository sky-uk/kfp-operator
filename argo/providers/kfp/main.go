package main

import (
	"context"
	"encoding/json"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/yalp/jsonpath"
	"io"
	"log"
	"os/exec"
)

type KfpProviderConfig struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

type KfpProvider struct {
}

func (kfpp KfpProvider) CreateExperiment(providerConfig KfpProviderConfig, experimentDefinition ExperimentDefinition, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "experiment", "create", experimentDefinition.Name)
	result, err := cmd.Output()

	if err != nil {
		return "", err
	}

	var output interface{}
	err = json.Unmarshal(result, &output)
	if err != nil {
		return "", err
	}

	id, err := jsonpath.Read(output, `$["ID"]`)
	if err != nil {
		return "", err
	}

	return id.(string), nil
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

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func main() {
	err := RunProviderApp[KfpProviderConfig](KfpProvider{})

	if err != nil {
		log.Fatal(err)
	}
}

func (kfpp KfpProvider) CreatePipeline(providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, pipelineFileName string, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload", "--pipeline-name", pipelineDefinition.Name, pipelineFileName)
	result, err := cmd.Output()

	if err != nil {
		return "", err
	}

	var output interface{}
	err = json.Unmarshal(result, &output)
	if err != nil {
		return "", err
	}

	id, err := jsonpath.Read(output, `$["Pipeline Details"]["Pipeline ID"]`)
	if err != nil {
		return "", err
	}

	return id.(string), nil
}

func (kfpp KfpProvider) UpdatePipeline(providerConfig KfpProviderConfig, pipelineDefinition PipelineDefinition, id string, pipelineFile string, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload-version", "--pipeline-version", pipelineDefinition.Version, "--pipeline-id", id, pipelineFile)
	result, err := cmd.Output()

	if err != nil {
		return "", err
	}

	var output interface{}
	err = json.Unmarshal(result, &output)
	if err != nil {
		return "", err
	}

	version, err := jsonpath.Read(output, `$["Version name"]`)
	return version.(string), nil
}

func (kfpp KfpProvider) DeletePipeline(providerConfig KfpProviderConfig, id string, _ context.Context) error {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "delete", id)
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
