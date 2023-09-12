package main

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/status-updater"
	"github.com/urfave/cli"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var EventingConstants = struct {
	RunCompletionEventParameter string
}{
	RunCompletionEventParameter: "run-completion-event",
}

func createK8sClient() (client.Client, error) {
	k8sConfig, err := common.K8sClientConfig()
	if err != nil {
		return nil, err
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(pipelinesv1.AddToScheme(scheme))

	return client.New(k8sConfig, client.Options{Scheme: scheme})
}

func main() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
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
				var rce common.RunCompletionEvent

				contents, err := os.ReadFile(c.GlobalString(EventingConstants.RunCompletionEventParameter))
				if err != nil {
					return err
				}

				err = json.Unmarshal(contents, &rce)
				if err != nil {
					return err
				}

				k8sClient, err := createK8sClient()
				if err != nil {
					return err
				}

				completer := status_updater.StatusUpdater{
					k8sClient,
				}
				return completer.UpdateStatus(ctx, rce)
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error(err, "failed to run")
		os.Exit(1)
	}
}
