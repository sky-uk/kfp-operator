package main

import (
	"context"
	configLoader "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	vai "github.com/sky-uk/kfp-operator/provider-service/vai/internal"
)

func main() {
	ctx := context.Background()
	config, err := configLoader.LoadConfig()
	if err != nil {
		panic(err)
	}
	vai.Start(ctx, *config)
}
