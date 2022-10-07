package main

import (
	"context"
	"encoding/json"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/yalp/jsonpath"
	"log"
	"os/exec"
)

type KfpProviderConfig struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

type KfpProvider struct {
}

func main() {
	err := RunProviderApp[KfpProviderConfig](KfpProvider{})

	if err != nil {
		log.Fatal(err)
	}
}

func (kfpp KfpProvider) CreatePipeline(providerConfig KfpProviderConfig, pipelineConfig PipelineConfig, pipelineFileName string, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload", "--pipeline-name", pipelineConfig.Name, pipelineFileName)
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

func (kfpp KfpProvider) UpdatePipeline(providerConfig KfpProviderConfig, pipelineConfig PipelineConfig, id string, pipelineFile string, _ context.Context) (string, error) {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "upload-version", "--pipeline-version", pipelineConfig.Version, "--pipeline-id", id, pipelineFile)
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

func (kfpp KfpProvider) DeletePipeline(providerConfig KfpProviderConfig, pipelineConfig PipelineConfig, id string, _ context.Context) error {
	cmd := exec.Command("kfp-ext", "--endpoint", providerConfig.Endpoint, "--output", "json", "pipeline", "delete", id)
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}
