//go:build unit

package mocks

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
)

type MockDeploymentManager struct {
	mock.Mock
}

func (m *MockDeploymentManager) Create(_ context.Context, deployment *appsv1.Deployment, provider *pipelinesv1.Provider) error {
	args := m.Called(deployment, provider)
	return args.Error(0)
}

func (m *MockDeploymentManager) Update(_ context.Context, old, new *appsv1.Deployment, provider *pipelinesv1.Provider) error {
	args := m.Called(old, new, provider)
	return args.Error(0)
}

func (m *MockDeploymentManager) Get(_ context.Context, provider *pipelinesv1.Provider) (*appsv1.Deployment, error) {
	args := m.Called(provider)
	var deployment *appsv1.Deployment
	if args.Get(0) != nil {
		deployment = args.Get(0).(*appsv1.Deployment)
	}
	return deployment, args.Error(1)
}

func (m *MockDeploymentManager) Equal(a, b *appsv1.Deployment) bool {
	args := m.Called(a, b)
	return args.Bool(0)
}

func (m *MockDeploymentManager) Construct(provider *pipelinesv1.Provider) (*appsv1.Deployment, error) {
	args := m.Called(provider)
	var deployment *appsv1.Deployment
	if args.Get(0) != nil {
		deployment = args.Get(0).(*appsv1.Deployment)
	}
	return deployment, args.Error(1)
}
