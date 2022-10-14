package base

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"os"
)

var ProviderConstants = struct {
	PipelineDefinitionParameter         string
	ExperimentDefinitionParameter       string
	RunConfigurationDefinitionParameter string
	ProviderConfigParameter             string
	PipelineIdParameter                 string
	ExperimentIdParameter               string
	RunConfigurationIdParameter         string
	PipelineFileParameter               string
	OutputParameter                     string
}{
	PipelineDefinitionParameter:         "pipeline-definition",
	ExperimentDefinitionParameter:       "experiment-definition",
	RunConfigurationDefinitionParameter: "runconfiguration-definition",
	ProviderConfigParameter:             "provider-config",
	PipelineIdParameter:                 "pipeline-id",
	ExperimentIdParameter:               "experiment-id",
	RunConfigurationIdParameter:         "runconfiguration-id",
	PipelineFileParameter:               "pipeline-file",
	OutputParameter:                     "out",
}

func RunProviderApp[Config any](provider Provider[Config]) error {
	providerConfigFlag := cli.StringFlag{
		Name:     ProviderConstants.ProviderConfigParameter,
		Required: true,
	}

	pipelineDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.PipelineDefinitionParameter,
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

	experimentDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.ExperimentDefinitionParameter,
		Required: true,
	}
	experimentIdFlag := cli.StringFlag{
		Name:     ProviderConstants.ExperimentIdParameter,
		Required: true,
	}

	runConfigurationDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.RunConfigurationDefinitionParameter,
		Required: true,
	}
	runConfigurationIdFlag := cli.StringFlag{
		Name:     ProviderConstants.RunConfigurationIdParameter,
		Required: true,
	}

	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name: "pipeline",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{providerConfigFlag, pipelineDefinitionFlag, pipelineFileFlag, outFlag},
					Action: func(c *cli.Context) error {
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}
						pipelineDefinition, err := loadPipelineDefinition(c)
						if err != nil {
							return err
						}

						id, err := provider.CreatePipeline(providerConfig, pipelineDefinition, pipelineFile, context.Background())
						if err != nil {
							return err
						}

						err = writeOutput(c, id)
						if err != nil {
							return err
						}

						fmt.Printf("Pipeline %s created\n", id)
						return nil
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, pipelineDefinitionFlag, pipelineFileFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						pipelineDefinition, err := loadPipelineDefinition(c)
						if err != nil {
							return err
						}
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						version, err := provider.UpdatePipeline(providerConfig, pipelineDefinition, id, pipelineFile, context.Background())
						if err != nil {
							return err
						}

						err = writeOutput(c, version)
						if err != nil {
							return err
						}

						fmt.Printf("Pipeline %s updated\n", id)
						return nil
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, pipelineDefinitionFlag, pipelineIdFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						err = provider.DeletePipeline(providerConfig, id, context.Background())
						if err != nil {
							return err
						}

						fmt.Printf("Pipeline %s deleted\n", id)
						return nil
					},
				},
			},
		},
		{
			Name: "experiment",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{providerConfigFlag, experimentDefinitionFlag, outFlag},
					Action: func(c *cli.Context) error {
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}
						experimentDefinition, err := loadExperimentDefinition(c)
						if err != nil {
							return err
						}

						id, err := provider.CreateExperiment(providerConfig, experimentDefinition, context.Background())
						if err != nil {
							return err
						}

						err = writeOutput(c, id)
						if err != nil {
							return err
						}

						fmt.Printf("Experiment %s created\n", id)
						return nil
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, experimentDefinitionFlag, experimentIdFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						err = provider.DeleteExperiment(providerConfig, id, context.Background())
						if err != nil {
							return err
						}

						fmt.Printf("Experiment %s deleted\n", id)
						return nil
					},
				},
			},
		},
		{
			Name: "runconfiguration",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{providerConfigFlag, runConfigurationDefinitionFlag, outFlag},
					Action: func(c *cli.Context) error {
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}
						runConfigurationDefinition, err := loadRunConfigurationDefinition(c)
						if err != nil {
							return err
						}

						id, err := provider.CreateRunConfiguration(providerConfig, runConfigurationDefinition, context.Background())
						if err != nil {
							return err
						}

						err = writeOutput(c, id)
						if err != nil {
							return err
						}

						fmt.Printf("RunConfiguration %s created\n", id)
						return nil
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, runConfigurationDefinitionFlag, runConfigurationIdFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						err = provider.DeleteRunConfiguration(providerConfig, id, context.Background())
						if err != nil {
							return err
						}

						fmt.Printf("RunConfiguration %s deleted\n", id)
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

func loadPipelineDefinition(c *cli.Context) (PipelineDefinition, error) {
	pipelineDefinition := PipelineDefinition{}

	pipelineDefinitionFile := c.String(ProviderConstants.PipelineDefinitionParameter)

	err := loadYamlFromFile(pipelineDefinitionFile, &pipelineDefinition)

	if err != nil {
		return pipelineDefinition, err
	}

	return pipelineDefinition, nil
}

func loadRunConfigurationDefinition(c *cli.Context) (RunConfigurationDefinition, error) {
	runConfigurationDefinition := RunConfigurationDefinition{}

	runConfigurationDefinitionFile := c.String(ProviderConstants.RunConfigurationDefinitionParameter)

	err := loadYamlFromFile(runConfigurationDefinitionFile, &runConfigurationDefinition)

	if err != nil {
		return runConfigurationDefinition, err
	}

	return runConfigurationDefinition, nil
}

func loadExperimentDefinition(c *cli.Context) (ExperimentDefinition, error) {
	experimentDefinition := ExperimentDefinition{}

	experimentDefinitionFile := c.String(ProviderConstants.ExperimentDefinitionParameter)

	err := loadYamlFromFile(experimentDefinitionFile, &experimentDefinition)

	if err != nil {
		return experimentDefinition, err
	}

	return experimentDefinition, nil
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
