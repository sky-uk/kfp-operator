package common

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
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
