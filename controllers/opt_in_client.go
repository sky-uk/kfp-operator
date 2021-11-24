package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OptInClient struct {
	client.Writer
	client.StatusClient
	Cached    client.Reader
	NonCached client.Reader
}

func NewOptInClient(mgr ctrl.Manager) OptInClient {
	managedClient := mgr.GetClient()

	return OptInClient{
		managedClient,
		managedClient,
		managedClient,
		mgr.GetAPIReader(),
	}
}
