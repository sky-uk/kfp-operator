//go:build decoupled
// +build decoupled

package run_completer

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type failingClient struct {
	err error
}

func NewFailingClient() client.Client {
	return failingClient{
		err: fmt.Errorf("failingClient error"),
	}
}


func (f failingClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return f.err
}

func (f failingClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return f.err
}

func (f failingClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return f.err
}

func (f failingClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
	return f.err
}

func (f failingClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return f.err
}

func (f failingClient) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return f.err
}

func (f failingClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) error {
	return f.err
}

func (f failingClient) Status() client.SubResourceWriter {
	return nil
}

func (f failingClient) SubResource(_ string) client.SubResourceClient {
	return nil
}

func (f failingClient) Scheme() *runtime.Scheme {
	return nil
}

func (f failingClient) RESTMapper() meta.RESTMapper {
	return nil
}


