//go:build unit

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockK8sClientReader struct{ mock.Mock }

func (m *MockK8sClientReader) Get(
	_ context.Context,
	key client.ObjectKey,
	obj client.Object,
	_ ...client.GetOption,
) error {
	args := m.Called(key, obj)
	return args.Error(0)
}

func (m *MockK8sClientReader) List(
	_ context.Context,
	list client.ObjectList,
	_ ...client.ListOption,
) error {
	args := m.Called(list)
	return args.Error(0)
}
