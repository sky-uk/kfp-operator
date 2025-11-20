package config

import (
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func K8sClientConfig() (k8sConfig *rest.Config, err error) {
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if _, err = os.Stat(kubeconfigPath); err == nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", "")
	}

	return
}

type OperatorConfig struct {
	Controller ControllerWrapper `yaml:"controller"`
	Spec       ConfigSpec        `yaml:"spec"`
}

type ControllerWrapper struct {
	GroupKindConcurrency    map[string]int `yaml:"groupKindConcurrency,omitempty"`
	MaxConcurrentReconciles int            `yaml:"maxConcurrentReconciles,omitempty"`
	CacheSyncTimeout        time.Duration  `yaml:"cacheSyncTimeout,omitempty"`
	RecoverPanic            *bool          `yaml:"recoverPanic,omitempty"`
	NeedLeaderElection      *bool          `yaml:"needLeaderElection,omitempty"`
}

type ConfigSpec struct {
	DefaultProvider        string                `yaml:"defaultProvider,omitempty"`
	DefaultProviderValues  DefaultProviderValues `yaml:"defaultProviderValues,omitempty"`
	DefaultTfxImage        string                `yaml:"defaultTfxImage,omitempty"`
	WorkflowTemplatePrefix string                `yaml:"workflowTemplatePrefix,omitempty"`
	WorkflowNamespace      string                `yaml:"workflowNamespace,omitempty"`
	Multiversion           bool                  `yaml:"multiversion,omitempty"`
	DefaultExperiment      string                `yaml:"defaultExperiment,omitempty"`
	RunCompletionTTL       *metav1.Duration      `yaml:"runCompletionTTL,omitempty"`
	RunCompletionFeed      ServiceConfiguration  `yaml:"runCompletionFeed,omitempty"`
}

type DefaultProviderValues struct {
	Labels               map[string]string  `yaml:"labels,omitempty"`
	Replicas             int                `yaml:"replicas,omitempty"`
	PodTemplateSpec      v1.PodTemplateSpec `yaml:"podTemplateSpec,omitempty"`
	ServiceContainerName string             `yaml:"serviceContainerName,omitempty"`
	ServicePort          int                `yaml:"servicePort,omitempty"`
	MetricsPort          int                `yaml:"metricsPort,omitempty"`
}

type ServiceConfiguration struct {
	Port      int        `json:"port,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

type Endpoint struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	Path string `json:"path,omitempty"`
}
