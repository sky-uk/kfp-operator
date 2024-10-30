package main

import (
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	stub "github.com/sky-uk/kfp-operator/argo/providers/stub/internal"
)

func main() {
	app := NewProviderApp[stub.StubProviderConfig]()
	provider := stub.StubProvider{}
	app.Run(provider)
}
