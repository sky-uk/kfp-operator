package main

import (
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	vai "github.com/sky-uk/kfp-operator/argo/providers/vai/internal"
)

func main() {
	app := NewProviderApp[vai.VAIProviderConfig]()
	app.Run(vai.VAIProvider{})
}
