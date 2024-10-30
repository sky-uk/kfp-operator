package main

import (
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	kfp "github.com/sky-uk/kfp-operator/argo/providers/kfp/internal"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	app := NewProviderApp[kfp.KfpProviderConfig]()
	app.Run(kfp.KfpProvider{})
}
