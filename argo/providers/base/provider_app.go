package base

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"os"
)

var ProviderConstants = struct {
	PipelineConfigParameter string
	ProviderConfigParameter string
	PipelineIdParameter     string
	PipelineFileParameter   string
	OutputParameter         string
}{
	PipelineConfigParameter: "pipeline-config",
	ProviderConfigParameter: "provider-config",
	PipelineIdParameter:     "pipeline-id",
	PipelineFileParameter:   "pipeline-file",
	OutputParameter:         "out",
}

func RunProviderApp[T any](provider Provider[T]) error {
	providerConfigFlag := cli.StringFlag{
		Name:     ProviderConstants.ProviderConfigParameter,
		Required: true,
	}

	pipelineConfigFlag := cli.StringFlag{
		Name:     ProviderConstants.PipelineConfigParameter,
		Required: true,
	}
	pipelineIdFlag := cli.StringFlag{
		Name:     ProviderConstants.PipelineIdParameter,
		Required: true,
	}

	pipelineFileFlag := cli.StringFlag{
		Name:     ProviderConstants.PipelineFileParameter,
		Required: true,
	}

	outFlag := cli.StringFlag{
		Name:     ProviderConstants.OutputParameter,
		Required: true,
	}

	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name: "pipeline",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{providerConfigFlag, pipelineConfigFlag, pipelineFileFlag, outFlag},
					Action: func(c *cli.Context) error {
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						providerConfig, err := loadProviderConfig[T](c)
						if err != nil {
							return err
						}
						pipelineConfig, err := loadPipelineConfig(c)
						if err != nil {
							return err
						}

						id, err := provider.CreatePipeline(providerConfig, pipelineConfig, pipelineFile, context.Background())
						if err != nil {
							return err
						}

						err = writeOutput(c, id)
						if err != nil {
							return err
						}

						return nil
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, pipelineConfigFlag, pipelineFileFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						pipelineConfig, err := loadPipelineConfig(c)
						if err != nil {
							return err
						}
						providerConfig, err := loadProviderConfig[T](c)
						if err != nil {
							return err
						}

						version, err := provider.UpdatePipeline(providerConfig, pipelineConfig, id, pipelineFile, context.Background())
						if err != nil {
							return err
						}

						err = writeOutput(c, version)
						if err != nil {
							return err
						}

						return nil
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, pipelineConfigFlag, pipelineIdFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						providerConfig, err := loadProviderConfig[T](c)
						if err != nil {
							return err
						}
						pipelineConfig, err := loadPipelineConfig(c)
						if err != nil {
							return err
						}

						err = provider.DeletePipeline(providerConfig, pipelineConfig, id, context.Background())
						if err != nil {
							return err
						}

						fmt.Printf("Pipeline %s deleted\n", id)
						return nil
					},
				},
			},
		},
	}

	return app.Run(os.Args)
}

func writeOutput(c *cli.Context, text string) error {
	return os.WriteFile(c.String(ProviderConstants.OutputParameter), []byte(text), 0644)
}

func loadYamlFromFile(fileName string, obj interface{}) error {
	content, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, obj)
	if err != nil {
		return err
	}

	return nil
}

func loadPipelineConfig(c *cli.Context) (PipelineConfig, error) {
	pipelineConfig := PipelineConfig{}

	pipelineConfigFile := c.String(ProviderConstants.PipelineConfigParameter)

	err := loadYamlFromFile(pipelineConfigFile, &pipelineConfig)

	if err != nil {
		return pipelineConfig, err
	}

	return pipelineConfig, nil
}

func loadProviderConfig[Config any](c *cli.Context) (Config, error) {

	providerConfig := new(Config)

	providerConfigFile := c.String(ProviderConstants.ProviderConfigParameter)
	err := loadYamlFromFile(providerConfigFile, providerConfig)

	if err != nil {
		return *providerConfig, err
	}

	return *providerConfig, nil
}
