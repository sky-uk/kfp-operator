package base

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"log"
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

func RunProviderApp[Config any](provider Provider[Config]) {
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

	logger, err := newLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}

	ctx := logr.NewContext(context.Background(), logger)
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
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}
						pipelineDefinition, err := loadFromParameter[PipelineDefinition](c, ProviderConstants.PipelineDefinitionParameter)
						if err != nil {
							return err
						}

						id, err := provider.CreatePipeline(ctx, providerConfig, pipelineDefinition, pipelineFile)

						logResult(ctx, "pipeline", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, pipelineDefinitionFlag, pipelineFileFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						pipelineDefinition, err := loadFromParameter[PipelineDefinition](c, ProviderConstants.PipelineDefinitionParameter)
						if err != nil {
							return err
						}
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdatePipeline(ctx, providerConfig, pipelineDefinition, id, pipelineFile)

						logResult(ctx, "pipeline", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, pipelineDefinitionFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}

						err = provider.DeletePipeline(ctx, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(ctx, "pipeline", "delete", id, updatedId, err)

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
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}
						runConfigurationDefinition, err := loadFromParameter[RunConfigurationDefinition](c, ProviderConstants.RunConfigurationDefinitionParameter)
						if err != nil {
							return err
						}

						id, err := provider.CreateRunConfiguration(ctx, providerConfig, runConfigurationDefinition)

						logResult(ctx, "runconfiguration", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, runConfigurationDefinitionFlag, runConfigurationIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}
						runConfigurationDefinition, err := loadFromParameter[RunConfigurationDefinition](c, ProviderConstants.RunConfigurationDefinitionParameter)
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateRunConfiguration(ctx, providerConfig, runConfigurationDefinition, id)

						logResult(ctx, "runconfiguration", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, runConfigurationDefinitionFlag, runConfigurationIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}

						err = provider.DeleteRunConfiguration(ctx, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(ctx, "runconfiguration", "delete", id, updatedId, err)

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
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}
						experimentDefinition, err := loadFromParameter[ExperimentDefinition](c, ProviderConstants.ExperimentDefinitionParameter)
						if err != nil {
							return err
						}

						id, err := provider.CreateExperiment(ctx, providerConfig, experimentDefinition)

						logResult(ctx, "experiment", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{providerConfigFlag, experimentDefinitionFlag, experimentIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}
						experimentDefinition, err := loadFromParameter[ExperimentDefinition](c, ProviderConstants.ExperimentDefinitionParameter)
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateExperiment(ctx, providerConfig, experimentDefinition, id)

						logResult(ctx, "experiment", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{providerConfigFlag, experimentDefinitionFlag, experimentIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.ExperimentIdParameter)
						providerConfig, err := loadFromParameter[Config](c, ProviderConstants.ProviderConfigParameter)
						if err != nil {
							return err
						}

						err = provider.DeleteExperiment(ctx, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(ctx, "experiment", "delete", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error(err, "failed to run provider app")
		os.Exit(1)
	}
}

func LoggerFromContext(ctx context.Context) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return logr.Discard()
	}

	return logger
}

func newLogger(logLevel zapcore.Level) (logr.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	zapLogger, err := config.Build()
	if err != nil {
		return logr.Discard(), err
	}

	return zapr.NewLogger(zapLogger.Named("main")), nil
}

func logResult(ctx context.Context, resourceType string, operation string, id string, updatedId string, err error) {
	logger := LoggerFromContext(ctx)

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

func loadFromParameter[T any](c *cli.Context, parameterName string) (T, error) {
	t := new(T)

	fileName := c.String(parameterName)
	content, err := os.ReadFile(fileName)
	if err == nil {
		err = yaml.Unmarshal(content, t)
	}

	return *t, err
}
