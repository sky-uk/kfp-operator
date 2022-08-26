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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1alpha3"
	pipelinesv1alpha2 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
	pipelinesv1alpha3 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	controllers "github.com/sky-uk/kfp-operator/controllers"
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

	utilruntime.Must(pipelinesv1alpha2.AddToScheme(scheme))
	utilruntime.Must(pipelinesv1alpha3.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
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
	ctrlConfig := configv1.KfpControllerConfig{}
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

	workflowFactory := pipelinescontrollers.WorkflowFactoryBase{
		Config: ctrlConfig.Workflows,
	}

	ec := pipelinescontrollers.K8sExecutionContext{
		Client:             client,
		Recorder:           mgr.GetEventRecorderFor("kfp-operator"),
		WorkflowRepository: workflowRepository,
	}

	if err = (&pipelinescontrollers.PipelineReconciler{
		EC: ec,
		StateHandler: pipelinescontrollers.PipelineStateHandler{
			WorkflowFactory: pipelinescontrollers.PipelineWorkflowFactory{
				WorkflowFactoryBase: workflowFactory,
			},
			WorkflowRepository: workflowRepository,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
		os.Exit(1)
	}
	if err = (&pipelinesv1alpha3.Pipeline{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Pipeline")
		os.Exit(1)
	}

	if err = (&pipelinescontrollers.RunConfigurationReconciler{
		EC: ec,
		StateHandler: pipelinescontrollers.RunConfigurationStateHandler{
			WorkflowFactory: pipelinescontrollers.RunConfigurationWorkflowFactory{
				WorkflowFactoryBase: workflowFactory,
			},
			WorkflowRepository: workflowRepository,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RunConfiguration")
		os.Exit(1)
	}
	//if err = (&pipelinesv1alpha3.RunConfiguration{}).SetupWebhookWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create webhook", "webhook", "RunConfiguration")
	//	os.Exit(1)
	//}

	if err = (&pipelinescontrollers.ExperimentReconciler{
		EC: ec,
		StateHandler: pipelinescontrollers.ExperimentStateHandler{
			WorkflowFactory: pipelinescontrollers.ExperimentWorkflowFactory{
				WorkflowFactoryBase: workflowFactory,
			},
			WorkflowRepository: workflowRepository,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Experiment")
		os.Exit(1)
	}
	//if err = (&pipelinesv1alpha3.Experiment{}).SetupWebhookWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create webhook", "webhook", "Experiment")
	//	os.Exit(1)
	//}

	//+kubebuilder:scaffold:builder

	if err = workflowRepository.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up WorkflowRepository")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
