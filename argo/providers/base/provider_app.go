package base

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/sky-uk/kfp-operator/providers/base/generic"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"os"
)

var ProviderConstants = struct {
	PipelineDefinitionParameter         string
	ExperimentDefinitionParameter       string
	RunDefinitionParameter              string
	RunConfigurationDefinitionParameter string
	ProviderConfigParameter             string
	PipelineIdParameter                 string
	ExperimentIdParameter               string
	RunConfigurationIdParameter         string
	RunIdParameter                      string
	PipelineFileParameter               string
	OutputParameter                     string
	EventsourceServerPortParameter      string
}{
	PipelineDefinitionParameter:         "pipeline-definition",
	ExperimentDefinitionParameter:       "experiment-definition",
	RunDefinitionParameter:              "run-definition",
	RunConfigurationDefinitionParameter: "runconfiguration-definition",
	ProviderConfigParameter:             "provider-config",
	PipelineIdParameter:                 "pipeline-id",
	ExperimentIdParameter:               "experiment-id",
	RunIdParameter:                      "run-id",
	RunConfigurationIdParameter:         "runconfiguration-id",
	PipelineFileParameter:               "pipeline-file",
	OutputParameter:                     "out",
	EventsourceServerPortParameter:      "port",
}

type ProviderApp[Config any] struct {
	Context context.Context
}

func NewProviderApp[Config any]() ProviderApp[Config] {
	logger, err := newLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}

	ctx := logr.NewContext(context.Background(), logger)
	return ProviderApp[Config]{
		Context: ctx,
	}
}

func (_ ProviderApp[Config]) LoadProviderConfig(c *cli.Context) (Config, error) {
	return LoadYamlFromFile[Config](c.GlobalString(ProviderConstants.ProviderConfigParameter))
}

func (providerApp ProviderApp[Config]) Run(provider Provider[Config], customCommands ...cli.Command) {
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

	runDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.RunDefinitionParameter,
		Required: true,
	}

	runConfigurationDefinitionFlag := cli.StringFlag{
		Name:     ProviderConstants.RunConfigurationDefinitionParameter,
		Required: true,
	}

	runIdFlag := cli.StringFlag{
		Name:     ProviderConstants.RunIdParameter,
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

	app.Flags = []cli.Flag{providerConfigFlag}
	app.Commands = []cli.Command{
		{
			Name: "pipeline",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{pipelineDefinitionFlag, pipelineFileFlag, outFlag},
					Action: func(c *cli.Context) error {
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						providerConfig, err := providerApp.LoadProviderConfig(c)
						if err != nil {
							return err
						}
						pipelineDefinition, err := LoadYamlFromFile[PipelineDefinition](c.String(ProviderConstants.PipelineDefinitionParameter))
						if err != nil {
							return err
						}

						id, err := provider.CreatePipeline(providerApp.Context, providerConfig, pipelineDefinition, pipelineFile)

						logResult(providerApp.Context, "pipeline", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{pipelineDefinitionFlag, pipelineFileFlag, pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						pipelineFile := c.String(ProviderConstants.PipelineFileParameter)
						providerConfig, err := providerApp.LoadProviderConfig(c)
						if err != nil {
							return err
						}
						pipelineDefinition, err := LoadYamlFromFile[PipelineDefinition](c.String(ProviderConstants.PipelineDefinitionParameter))
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdatePipeline(providerApp.Context, providerConfig, pipelineDefinition, id, pipelineFile)

						logResult(providerApp.Context, "pipeline", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{pipelineIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.PipelineIdParameter)
						providerConfig, err := providerApp.LoadProviderConfig(c)
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
			Name: "runconfiguration",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Flags: []cli.Flag{runConfigurationDefinitionFlag, outFlag},
					Action: func(c *cli.Context) error {
						providerConfig, err := providerApp.LoadProviderConfig(c)
						if err != nil {
							return err
						}
						runConfigurationDefinition, err := LoadYamlFromFile[RunConfigurationDefinition](c.String(ProviderConstants.RunConfigurationDefinitionParameter))
						if err != nil {
							return err
						}
						id, err := provider.CreateRunConfiguration(providerApp.Context, providerConfig, runConfigurationDefinition)

						logResult(providerApp.Context, "runconfiguration", "create", "", id, err)

						return writeOutput(c, id, err)
					},
				},
				{
					Name:  "update",
					Flags: []cli.Flag{runConfigurationDefinitionFlag, runConfigurationIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := providerApp.LoadProviderConfig(c)
						if err != nil {
							return err
						}
						runConfigurationDefinition, err := LoadYamlFromFile[RunConfigurationDefinition](c.String(ProviderConstants.RunConfigurationDefinitionParameter))
						if err != nil {
							return err
						}

						updatedId, err := provider.UpdateRunConfiguration(providerApp.Context, providerConfig, runConfigurationDefinition, id)

						logResult(providerApp.Context, "runconfiguration", "update", id, updatedId, err)

						return writeOutput(c, updatedId, err)
					},
				},
				{
					Name:  "delete",
					Flags: []cli.Flag{runConfigurationIdFlag, outFlag},
					Action: func(c *cli.Context) error {
						id := c.String(ProviderConstants.RunConfigurationIdParameter)
						providerConfig, err := providerApp.LoadProviderConfig(c)
						if err != nil {
							return err
						}

						err = provider.DeleteRunConfiguration(providerApp.Context, providerConfig, id)
						updatedId := ""
						if err != nil {
							updatedId = id
						}

						logResult(providerApp.Context, "runconfiguration", "delete", id, updatedId, err)

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
						providerConfig, err := providerApp.LoadProviderConfig(c)
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
						providerConfig, err := providerApp.LoadProviderConfig(c)
						if err != nil {
							return err
						}

						err = provider.DeleteRunConfiguration(providerApp.Context, providerConfig, id)
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
						providerConfig, err := providerApp.LoadProviderConfig(c)
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
						providerConfig, err := providerApp.LoadProviderConfig(c)
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
						providerConfig, err := providerApp.LoadProviderConfig(c)
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
		{
			Name: "eventsource-server",
			Flags: []cli.Flag{cli.StringFlag{
				Name:     ProviderConstants.EventsourceServerPortParameter,
				Required: true,
			}},
			Action: func(c *cli.Context) error {
				logger := LoggerFromContext(providerApp.Context)
				providerConfig, err := providerApp.LoadProviderConfig(c)

				lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", c.Int(ProviderConstants.EventsourceServerPortParameter)))
				if err != nil {
					logger.Error(err, "failed to listen")
					os.Exit(1)
				}

				s := grpc.NewServer()

				eventingServer, err := provider.EventingServer(providerApp.Context, providerConfig)
				if err != nil {
					logger.Error(err, "failed to create eventing server")
					os.Exit(1)
				}

				generic.RegisterEventingServer(s, eventingServer)

				logger.Info(fmt.Sprintf("server listening at %s", lis.Addr()))
				if err := s.Serve(lis); err != nil {
					logger.Error(err, "failed to serve")
					os.Exit(1)
				}

				return nil
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
		LoggerFromContext(providerApp.Context).Error(err, "failed to run provider app")
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
