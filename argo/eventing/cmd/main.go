package main

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"github.com/sky-uk/kfp-operator/argo/eventing"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: common code to refactor
var EventingConstants = struct {
	RunCompletionEventParameter string
}{
	RunCompletionEventParameter: "run-completion-event",
}

func LoggerFromContext(ctx context.Context) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return logr.Discard()
	}

	return logger
}

// TODO: common code to refactor
func newLogger(logLevel zapcore.Level) (logr.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	zapLogger, err := config.Build()
	if err != nil {
		return logr.Discard(), err
	}

	return zapr.NewLogger(zapLogger.Named("main")), nil
}

// TODO: common code to refactor
func createK8sClient() (client.Client, error) {
	var k8sConfig *rest.Config
	var err error

	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfigPath); err == nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", "")
	}
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(pipelinesv1.AddToScheme(scheme))

	return client.New(k8sConfig, client.Options{Scheme: scheme})
}

func main() {
	logger, err := newLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}

	ctx := logr.NewContext(context.Background(), logger)

	app := cli.NewApp()

	runCompletionEventFlag := cli.StringFlag{
		Name:     EventingConstants.RunCompletionEventParameter,
		Required: true,
	}

	app.Flags = []cli.Flag{runCompletionEventFlag}

	app.Commands = []cli.Command{
		{
			Name: "complete-run",
			Action: func(c *cli.Context) error {
				var rce eventing.RunCompletionEvent

				contents, err := os.ReadFile(c.GlobalString(EventingConstants.RunCompletionEventParameter))
				if err != nil {
					return err
				}

				err = yaml.Unmarshal(contents, &rce)
				if err != nil {
					return err
				}

				k8sClient, err := createK8sClient()
				if err != nil {
					return err
				}

				completer := eventing.RunCompleter{
					k8sClient,
				}
				return completer.CompleteRun(ctx, rce)
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error(err, "failed to run provider app")
		os.Exit(1)
	}
}
