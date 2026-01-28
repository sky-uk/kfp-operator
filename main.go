/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/controllers/mcp"
	"github.com/sky-uk/kfp-operator/controllers/webhook"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/internal/metrics"
	"k8s.io/apimachinery/pkg/util/yaml"
	runtimeMetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	pipelineshubalpha5 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	pipelineshubalpha6 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers"
	pipelinescontrollers "github.com/sky-uk/kfp-operator/controllers/pipelines"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	//+kubebuilder:scaffold:imports

	"github.com/sky-uk/kfp-operator/external"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(pipelineshubalpha5.AddToScheme(scheme))
	utilruntime.Must(pipelineshubalpha6.AddToScheme(scheme))
	utilruntime.Must(pipelineshub.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	utilruntime.Must(external.InitSchemes(scheme))
	_, err := metrics.InitMeterProvider("kfp-operator-controller-manager", prometheus.WithRegisterer(runtimeMetrics.Registry))
	if err != nil {
		setupLog.Error(err, "failed to initialise metrics")
		os.Exit(1)
	}
}

func main() {
	opts := zap.Options{
		Development: false,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var configFile string

	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	var err error
	ctrlConfig := config.OperatorConfig{}
	options := ctrl.Options{Scheme: scheme, HealthProbeBindAddress: ":8081"}

	if configFile != "" {
		bytes, err := os.ReadFile(configFile)
		if err != nil {
			setupLog.Error(err, "unable to read the config file", "path", configFile)
			os.Exit(1)
		}
		if err = yaml.Unmarshal(bytes, &ctrlConfig); err != nil {
			setupLog.Error(err, "unable to parse the config file", "path", configFile, "content", string(bytes))
			os.Exit(1)
		}
		options.Controller = ctrlConfig.Controller.ToController()
	}

	// TODO: This is temporary whilst have conversion from v1alpha5/6 to v1beta1, this is to be removed once v1alpha6 is removed.
	pipelineshubalpha5.DefaultProvider = ctrlConfig.Spec.DefaultProvider
	pipelineshubalpha5.DefaultProviderNamespace = ctrlConfig.Spec.WorkflowNamespace
	pipelineshubalpha6.DefaultProviderNamespace = ctrlConfig.Spec.WorkflowNamespace
	pipelineshubalpha5.DefaultTfxImage = ctrlConfig.Spec.DefaultTfxImage
	pipelineshubalpha6.DefaultTfxImage = ctrlConfig.Spec.DefaultTfxImage

	var mgr ctrl.Manager

	mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	client := controllers.NewOptInClient(mgr)

	workflowRepository := pipelinescontrollers.WorkflowRepositoryImpl{
		Client: client,
		Config: ctrlConfig.Spec,
		Scheme: mgr.GetScheme(),
	}

	ec := pipelinescontrollers.K8sExecutionContext{
		Client:             client,
		Recorder:           mgr.GetEventRecorderFor("kfp-operator"),
		Scheme:             mgr.GetScheme(),
		WorkflowRepository: workflowRepository,
	}

	if err := pipelinescontrollers.NewPipelineReconciler(
		ec,
		workflowRepository,
		ctrlConfig.Spec,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewRunReconciler(
		ec,
		workflowRepository,
		ctrlConfig.Spec,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Run")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewRunScheduleReconciler(
		ec,
		workflowRepository,
		ctrlConfig.Spec,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RunSchedule")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewExperimentReconciler(
		ec,
		workflowRepository,
		ctrlConfig.Spec,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Experiment")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewRunConfigurationReconciler(
		ec,
		ctrlConfig.Spec,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RunConfiguration")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewProviderReconciler(ec, ctrlConfig.Spec).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Provider")
		os.Exit(1)
	}

	if ctrlConfig.Spec.Multiversion {
		if err = pipelineshub.NewPipelineValidatorWebhook(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Pipeline")
			os.Exit(1)
		}
		if err = (&pipelineshub.Experiment{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Experiment")
			os.Exit(1)
		}
		if err = pipelineshub.NewRunConfigurationValidatorWebhook(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RunConfiguration")
			os.Exit(1)
		}
		if err = (&pipelineshub.RunSchedule{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RunSchedule")
			os.Exit(1)
		}
		if err = (&pipelineshub.Provider{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Provider")
			os.Exit(1)
		}
	}

	// Resources that have a validation webhook in addition to conversion
	if err = pipelineshub.NewRunValidatorWebhook(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Run")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	setupLog.Info("starting run completion feed")

	handlers := lo.Map(
		ctrlConfig.Spec.RunCompletionFeed.Endpoints,
		func(endpoint config.Endpoint, _ int) webhook.RunCompletionEventHandler {
			return webhook.NewRunCompletionEventTrigger(ctx, endpoint)
		},
	)
	statusUpdater, err := webhook.NewStatusUpdater(ctx, scheme)
	if err != nil {
		setupLog.Error(err, "unable to create status updater")
		os.Exit(1)
	}
	handlers = append(handlers, statusUpdater)
	rcf, err := webhook.NewObservedRunCompletionFeed(
		client.NonCached,
		handlers,
	)
	if err != nil {
		setupLog.Error(err, "unable to create run completion feed")
		os.Exit(1)
	}
	go func() {
		http.HandleFunc("/events", rcf.HandleEvent(ctx))
		err := http.ListenAndServe(fmt.Sprintf(":%d", ctrlConfig.Spec.RunCompletionFeed.Port), nil)
		if err != nil {
			setupLog.Error(err, "problem starting run completion feed")
			os.Exit(1)
		}
	}()

	mcpServer := mcp.MCPServer{
		Cache: mgr.GetCache(),
	}
	runnable := mcp.Runnable{Server: &mcpServer}
	err = runnable.Start(ctx)

	if err != nil {
		setupLog.Error(err, "problem starting mcp server")
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem starting manager")
		os.Exit(1)
	}
}
