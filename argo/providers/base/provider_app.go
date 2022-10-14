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

	runConfigurationDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.RunConfigurationDefinitionParameter,
		Required: true,
	}

	runConfigurationIdFlag := cli.StringFlag{
		Name:     ProviderConstants.RunConfigurationIdParameter,
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

						printResult("pipeline", "create", "", id, err)

						return writeOutput(c, id, err)
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

						updatedId, err := provider.UpdatePipeline(providerConfig, pipelineDefinition, id, pipelineFile, context.Background())

						printResult("pipeline", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, pipelineDefinitionFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						err = provider.DeletePipeline(providerConfig, id, context.Background())
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						printResult("pipeline", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
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

						printResult("runconfiguration", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, runConfigurationDefinitionFlag, runConfigurationIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}
						runConfigurationDefinition, err := loadRunConfigurationDefinition(c)
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateRunConfiguration(providerConfig, runConfigurationDefinition, id, context.Background())

						printResult("runconfiguration", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, runConfigurationDefinitionFlag, runConfigurationIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						err = provider.DeleteRunConfiguration(providerConfig, id, context.Background())
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						printResult("runconfiguration", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
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

						printResult("experiment", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, experimentDefinitionFlag, experimentIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}
						experimentDefinition, err := loadExperimentDefinition(c)
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateExperiment(providerConfig, experimentDefinition, id, context.Background())

						printResult("experiment", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, experimentDefinitionFlag, experimentIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := loadProviderConfig[Config](c)
						if err != nil {
							return err
						}

						err = provider.DeleteExperiment(providerConfig, id, context.Background())
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						printResult("experiment", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
			},
		},
	}

	return app.Run(os.Args)
}

func printResult(resourceType string, operation string, id string, updatedId string, err error) {
	//TODO: use logging
	if err == nil {
		fmt.Printf("%s %s succeeded. Id: %s -> %s\n", resourceType, operation, id, updatedId)
	} else {
		fmt.Printf("%s %s failed. Id: %s -> %s. Error: %e\n", resourceType, operation, id, updatedId, err)
	}
}

func writeOutput(c *cli.Context, id string, err error) error {
	errMessage := ""
	if err != nil {
		errMessage = err.Error()
	}

	output := Output{
		Id:            id,
		ProviderError: errMessage,
	}

	outputYaml, err := yaml.Marshal(&output)
	if err != nil {
		return err
	}

	return os.WriteFile(c.String(ProviderConstants.OutputParameter), outputYaml, 0644)
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
