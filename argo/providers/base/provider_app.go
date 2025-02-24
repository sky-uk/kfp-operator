package base

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/urfave/cli"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var ProviderConstants = struct {
	PipelineDefinitionParameter    string
	ExperimentDefinitionParameter  string
	RunDefinitionParameter         string
	RunScheduleDefinitionParameter string
	ProviderParameter              string
	PipelineIdParameter            string
	ExperimentIdParameter          string
	RunScheduleIdParameter         string
	RunIdParameter                 string
	PipelineFileParameter          string
	OutputParameter                string
	Namespace                      string
}{
	PipelineDefinitionParameter:    "pipeline-definition",
	ExperimentDefinitionParameter:  "experiment-definition",
	RunDefinitionParameter:         "run-definition",
	RunScheduleDefinitionParameter: "runschedule-definition",
	ProviderParameter:              "provider",
	PipelineIdParameter:            "pipeline-id",
	ExperimentIdParameter:          "experiment-id",
	RunIdParameter:                 "run-id",
	RunScheduleIdParameter:         "runschedule-id",
	PipelineFileParameter:          "pipeline-file",
	OutputParameter:                "out",
	Namespace:                      "namespace",
}

type ProviderApp[Config any] struct {
	Context context.Context
}

func NewProviderApp[Config any]() ProviderApp[Config] {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}

	ctx := logr.NewContext(context.Background(), logger)
	return ProviderApp[Config]{
		Context: ctx,
	}
}

func (providerApp ProviderApp[Config]) Run(provider Provider[Config], customCommands ...cli.Command) {

	providerFlag := cli.StringFlag{
		Name:     ProviderConstants.ProviderParameter,
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

	runDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.RunDefinitionParameter,
		Required: true,
	}

	runScheduleDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.RunScheduleDefinitionParameter,
		Required: true,
	}

	runIdFlag := cli.StringFlag{
		Name:     ProviderConstants.RunIdParameter,
		Required: true,
	}

	runScheduleIdFlag := cli.StringFlag{
		Name:     ProviderConstants.RunScheduleIdParameter,
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

	app.Flags = []cli.Flag{providerFlag}
	app.Commands = []cli.Command{
		{
			Name: "pipeline",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{pipelineDefinitionFlag, pipelineFileFlag, outFlag},
					Action: func(c *cli.Context) error {
						pipelineFilePath := c.String(ProviderConstants.PipelineFileParameter)
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						pipelineDefinition, err := LoadYamlFromFile[PipelineDefinition](c.String(ProviderConstants.PipelineDefinitionParameter))
						if err != nil {
							return err
						}

						id, err := provider.CreatePipeline(providerApp.Context, providerConfig, pipelineDefinition, pipelineFilePath)

						logResult(providerApp.Context, "pipeline", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{pipelineDefinitionFlag, pipelineFileFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						pipelineFilePath := c.String(ProviderConstants.PipelineFileParameter)
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						pipelineDefinition, err := LoadYamlFromFile[PipelineDefinition](c.String(ProviderConstants.PipelineDefinitionParameter))
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdatePipeline(providerApp.Context, providerConfig, pipelineDefinition, id, pipelineFilePath)

						logResult(providerApp.Context, "pipeline", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						err = provider.DeletePipeline(providerApp.Context, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(providerApp.Context, "pipeline", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
			},
		},
		{
			Name: "runschedule",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{runScheduleDefinitionFlag, outFlag},
					Action: func(c *cli.Context) error {
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						runScheduleDefinition, err := LoadYamlFromFile[RunScheduleDefinition](c.String(ProviderConstants.RunScheduleDefinitionParameter))
						if err != nil {
							return err
						}

						id, err := provider.CreateRunSchedule(providerApp.Context, providerConfig, runScheduleDefinition)

						logResult(providerApp.Context, "runschedule", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{runScheduleDefinitionFlag, runScheduleIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunScheduleIdParameter)

						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						runScheduleDefinition, err := LoadYamlFromFile[RunScheduleDefinition](c.String(ProviderConstants.RunScheduleDefinitionParameter))
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateRunSchedule(providerApp.Context, providerConfig, runScheduleDefinition, id)

						logResult(providerApp.Context, "runschedule", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{runScheduleIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunScheduleIdParameter)

						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						err = provider.DeleteRunSchedule(providerApp.Context, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(providerApp.Context, "runschedule", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
			},
		},
		{
			Name: "run",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{runDefinitionFlag, outFlag},
					Action: func(c *cli.Context) error {
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						runDefinition, err := LoadYamlFromFile[RunDefinition](c.String(ProviderConstants.RunDefinitionParameter))
						if err != nil {
							return err
						}

						id, err := provider.CreateRun(providerApp.Context, providerConfig, runDefinition)

						logResult(providerApp.Context, "run", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{runIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunIdParameter)
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						err = provider.DeleteRun(providerApp.Context, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(providerApp.Context, "run", "delete", id, updatedId, err)

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
					Flags: []cli.Flag{experimentDefinitionFlag, outFlag},
					Action: func(c *cli.Context) error {
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						experimentDefinition, err := LoadYamlFromFile[ExperimentDefinition](c.String(ProviderConstants.ExperimentDefinitionParameter))
						if err != nil {
							return err
						}

						id, err := provider.CreateExperiment(providerApp.Context, providerConfig, experimentDefinition)

						logResult(providerApp.Context, "experiment", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{experimentDefinitionFlag, experimentIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						experimentDefinition, err := LoadYamlFromFile[ExperimentDefinition](c.String(ProviderConstants.ExperimentDefinitionParameter))
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateExperiment(providerApp.Context, providerConfig, experimentDefinition, id)

						logResult(providerApp.Context, "experiment", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{experimentIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := LoadJsonFromFile[Config](c.GlobalString(ProviderConstants.ProviderParameter))
						if err != nil {
							return err
						}

						err = provider.DeleteExperiment(providerApp.Context, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(providerApp.Context, "experiment", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
			},
		},
	}

	if len(customCommands) > 0 {
		app.Commands = append(app.Commands, cli.Command{
			Name:        "custom",
			Subcommands: customCommands,
		})
	}

	if err := app.Run(os.Args); err != nil {
		common.LoggerFromContext(providerApp.Context).Error(err, "failed to run provider app")
		os.Exit(1)
	}
}

func logResult(ctx context.Context, resourceType string, operation string, id string, updatedId string, err error) {
	logger := common.LoggerFromContext(ctx)

	var msg string
	if err == nil {
		msg = "operation succeeded"
	} else {
		msg = "operation failed"
	}

	logger.Info(msg, "operation", operation, "resource type", resourceType, "id", id, "updated id", updatedId)
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

func LoadYamlFromFile[T any](fileName string) (T, error) {
	t := new(T)

	contents, err := os.ReadFile(fileName)
	if err == nil {
		err = yaml.Unmarshal(contents, t)
	}

	return *t, err
}

func LoadJsonFromFile[T any](fileName string) (T, error) {
	t := new(T)

	contents, err := os.ReadFile(fileName)
	if err == nil {
		err = json.Unmarshal(contents, t)
	}

	return *t, err
}
