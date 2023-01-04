package main

import (
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/stub"
)

func main() {
	app := NewProviderApp[stub.StubProviderConfig]()
	provider := stub.StubProvider{}
	app.Run(provider)
}
