//go:build unit

package mocks

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

type MockServiceManager struct {
	mock.Mock
}

func (m *MockServiceManager) Create(_ context.Context, service *corev1.Service, provider *pipelinesv1.Provider) error {
	args := m.Called(service, provider)
	return args.Error(0)
}

func (m *MockServiceManager) Delete(_ context.Context, service *corev1.Service) error {
	args := m.Called(service)
	return args.Error(0)
}

func (m *MockServiceManager) Get(_ context.Context, provider *pipelinesv1.Provider) (*corev1.Service, error) {
	args := m.Called(provider)
	var service *corev1.Service
	if args.Get(0) != nil {
		service = args.Get(0).(*corev1.Service)
	}
	return service, args.Error(1)
}

func (m *MockServiceManager) Equal(a, b *corev1.Service) bool {
	args := m.Called(a, b)
	return args.Bool(0)
}

func (m *MockServiceManager) Construct(provider *pipelinesv1.Provider) *corev1.Service {
	args := m.Called(provider)
	return args.Get(0).(*corev1.Service)
}
