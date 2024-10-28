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
	"os"

	"github.com/sky-uk/kfp-operator/controllers/webhook"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers"
	pipelinescontrollers "github.com/sky-uk/kfp-operator/controllers/pipelines"

	//+kubebuilder:scaffold:imports

	"github.com/sky-uk/kfp-operator/external"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(pipelinesv1.AddToScheme(scheme))
	utilruntime.Must(config.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	utilruntime.Must(external.InitSchemes(scheme))
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	var err error
	ctrlConfig := config.KfpControllerConfig{}
	options := ctrl.Options{Scheme: scheme}

	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	var mgr ctrl.Manager

	mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	client := controllers.NewOptInClient(mgr)

	workflowRepository := pipelinescontrollers.WorkflowRepositoryImpl{
		Client: client,
		Scheme: mgr.GetScheme(),
	}

	ec := pipelinescontrollers.K8sExecutionContext{
		Client:             client,
		Recorder:           mgr.GetEventRecorderFor("kfp-operator"),
		Scheme:             mgr.GetScheme(),
		WorkflowRepository: workflowRepository,
	}

	if err := pipelinescontrollers.NewPipelineReconciler(ec, workflowRepository, ctrlConfig.Spec).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewRunReconciler(ec, workflowRepository, ctrlConfig.Spec).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Run")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewRunScheduleReconciler(ec, workflowRepository, ctrlConfig.Spec).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RunSchedule")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewExperimentReconciler(ec, workflowRepository, ctrlConfig.Spec).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Experiment")
		os.Exit(1)
	}

	if err = pipelinescontrollers.NewRunConfigurationReconciler(ec, scheme, ctrlConfig.Spec).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RunConfiguration")
		os.Exit(1)
	}

	if ctrlConfig.Spec.Multiversion {
		if err = (&pipelinesv1.Pipeline{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Pipeline")
			os.Exit(1)
		}
		if err = (&pipelinesv1.Experiment{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Experiment")
			os.Exit(1)
		}
		if err = (&pipelinesv1.RunConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RunConfiguration")
			os.Exit(1)
		}
		if err = (&pipelinesv1.RunSchedule{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RunSchedule")
			os.Exit(1)
		}
		if err = (&pipelinesv1.Provider{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Provider")
			os.Exit(1)
		}
	}

	// Resources that have a validation webhook in addition to conversion
	if err = (&pipelinesv1.Run{}).SetupWebhookWithManager(mgr); err != nil {
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
	rcf := webhook.NewRunCompletionFeed(ctx, client.NonCached, ctrlConfig.Spec.RunCompletionFeed.Endpoints)
	go func() {
		err = rcf.Start(ctrlConfig.Spec.RunCompletionFeed.Port)
		if err != nil {
			setupLog.Error(err, "problem starting run completion feed")
			os.Exit(1)
		}
	}()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem starting manager")
		os.Exit(1)
	}
}
