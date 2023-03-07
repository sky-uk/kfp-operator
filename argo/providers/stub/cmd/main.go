package main

import (
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/stub"
)

func main() {
	app := NewProviderApp[stub.StubProviderConfig]()
	provider := stub.StubProvider{}
	app.Run(provider)
}
