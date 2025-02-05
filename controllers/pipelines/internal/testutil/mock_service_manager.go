//go:build unit

package testutil

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	corev1 "k8s.io/api/core/v1"
)

type MockServiceManager struct {
	CreateFunc    func() error
	DeleteFunc    func() error
	GetFunc       func() (*corev1.Service, error)
	EqualFunc     func() bool
	ConstructFunc func() (*corev1.Service, error)
}

func (m *MockServiceManager) Create(_ context.Context, _ *corev1.Service, _ *pipelinesv1.Provider) error {
	if m.CreateFunc != nil {
		return m.CreateFunc()
	}
	return nil
}

func (m *MockServiceManager) Delete(_ context.Context, _ *corev1.Service) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc()
	}
	return nil
}

func (m *MockServiceManager) Get(_ context.Context, _ *pipelinesv1.Provider) (*corev1.Service, error) {
	if m.GetFunc != nil {
		return m.GetFunc()
	}
	return &corev1.Service{}, nil
}

func (m *MockServiceManager) Equal(_, _ *corev1.Service) bool {
	if m.EqualFunc != nil {
		return m.EqualFunc()
	}
	return true
}

func (m *MockServiceManager) Construct(_ *pipelinesv1.Provider) (*corev1.Service, error) {
	if m.ConstructFunc != nil {
		return m.ConstructFunc()
	}
	return &corev1.Service{}, nil
}
