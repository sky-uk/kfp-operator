// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/sky-uk/kfp-operator/argo/providers/vai (interfaces: PipelineJobClient)

// Package vai is a generated GoMock package.
package vai

import (
	context "context"
	reflect "reflect"

	aiplatformpb "cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	gomock "github.com/golang/mock/gomock"
	gax "github.com/googleapis/gax-go/v2"
)

// MockPipelineJobClient is a mock of PipelineJobClient interface.
type MockPipelineJobClient struct {
	ctrl     *gomock.Controller
	recorder *MockPipelineJobClientMockRecorder
}

// MockPipelineJobClientMockRecorder is the mock recorder for MockPipelineJobClient.
type MockPipelineJobClientMockRecorder struct {
	mock *MockPipelineJobClient
}

// NewMockPipelineJobClient creates a new mock instance.
func NewMockPipelineJobClient(ctrl *gomock.Controller) *MockPipelineJobClient {
	mock := &MockPipelineJobClient{ctrl: ctrl}
	mock.recorder = &MockPipelineJobClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPipelineJobClient) EXPECT() *MockPipelineJobClientMockRecorder {
	return m.recorder
}

// GetPipelineJob mocks base method.
func (m *MockPipelineJobClient) GetPipelineJob(arg0 context.Context, arg1 *aiplatformpb.GetPipelineJobRequest, arg2 ...gax.CallOption) (*aiplatformpb.PipelineJob, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPipelineJob", varargs...)
	ret0, _ := ret[0].(*aiplatformpb.PipelineJob)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPipelineJob indicates an expected call of GetPipelineJob.
func (mr *MockPipelineJobClientMockRecorder) GetPipelineJob(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPipelineJob", reflect.TypeOf((*MockPipelineJobClient)(nil).GetPipelineJob), varargs...)
}
