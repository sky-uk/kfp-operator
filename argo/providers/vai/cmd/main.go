package main

import (
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/vai"
)

func main() {
	app := NewProviderApp[vai.VAIProviderConfig]()
	app.Run(vai.VAIProvider{})
}
