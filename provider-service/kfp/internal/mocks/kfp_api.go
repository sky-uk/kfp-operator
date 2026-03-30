//go:build decoupled || unit

package mocks

import (
	"context"

	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/stretchr/testify/mock"
)

type MockKfpApi struct {
	mock.Mock
}

func (m *MockKfpApi) GetResourceReferences(
	_ context.Context,
	runId string,
) (resource.References, error) {
	args := m.Called(runId)
	var refs resource.References
	if arg0 := args.Get(0); arg0 != nil {
		refs = arg0.(resource.References)
	}
	return refs, args.Error(1)
}
