package main

import (
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	. "github.com/sky-uk/kfp-operator/argo/providers/kfp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	app := NewProviderApp[KfpProviderConfig]()
	app.Run(KfpProvider{})
}
