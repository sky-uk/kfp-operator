package main

import (
	"context"
	configLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	kfp "github.com/sky-uk/kfp-operator/provider-service/kfp/internal"
)

func main() {
	ctx := context.Background()
	config, err := configLoader.LoadConfig()
	if err != nil {
		panic(err)
	}
	kfp.Start(ctx, *config)
}
