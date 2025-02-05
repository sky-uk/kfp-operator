//go:build unit

package testutil

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	appsv1 "k8s.io/api/apps/v1"
)

type MockDeploymentManager struct {
	CreateFunc    func() error
	UpdateFunc    func() error
	GetFunc       func() (*appsv1.Deployment, error)
	EqualFunc     func() bool
	ConstructFunc func() (*appsv1.Deployment, error)
}

func (m *MockDeploymentManager) Create(_ context.Context, _ *appsv1.Deployment, _ *pipelinesv1.Provider) error {
	if m.CreateFunc != nil {
		return m.CreateFunc()
	}
	return nil
}

func (m *MockDeploymentManager) Update(_ context.Context, _ *appsv1.Deployment, _ *appsv1.Deployment, _ *pipelinesv1.Provider) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc()
	}
	return nil
}

func (m *MockDeploymentManager) Get(_ context.Context, _ *pipelinesv1.Provider) (*appsv1.Deployment, error) {
	if m.GetFunc != nil {
		return m.GetFunc()
	}
	return &appsv1.Deployment{}, nil
}

func (m *MockDeploymentManager) Equal(_, _ *appsv1.Deployment) bool {
	if m.EqualFunc != nil {
		return m.EqualFunc()
	}
	return true
}

func (m *MockDeploymentManager) Construct(_ *pipelinesv1.Provider) (*appsv1.Deployment, error) {
	if m.ConstructFunc != nil {
		return m.ConstructFunc()
	}
	return &appsv1.Deployment{}, nil
}
