package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	System SystemConfig `yaml:"system"`
	Spec   ConfigSpec   `yaml:"spec"`
}

type SystemConfig struct {
	LeaderElection LeaderElectionConfig `yaml:"leaderElection"`
}

type LeaderElectionConfig struct {
	Enabled bool   `yaml:"enabled"`
	Id      string `yaml:"id"`
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
	RunCompletionFeed      ServiceConfig         `yaml:"runCompletionFeed,omitempty"`
}

type DefaultProviderValues struct {
	Labels               map[string]string  `yaml:"labels,omitempty"`
	Replicas             int                `yaml:"replicas,omitempty"`
	PodTemplateSpec      v1.PodTemplateSpec `yaml:"podTemplateSpec,omitempty"`
	ServiceContainerName string             `yaml:"serviceContainerName,omitempty"`
	ServicePort          int                `yaml:"servicePort,omitempty"`
	MetricsPort          int                `yaml:"metricsPort,omitempty"`
}

type ServiceConfig struct {
	Port      int        `yaml:"port,omitempty"`
	Endpoints []Endpoint `yaml:"endpoints,omitempty"`
}

type Endpoint struct {
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
	Path string `yaml:"path,omitempty"`
}

func (e Endpoint) URL() string {
	return fmt.Sprintf("%s:%d%s", e.Host, e.Port, e.Path)
}
