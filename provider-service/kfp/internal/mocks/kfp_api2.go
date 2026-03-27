//go:build decoupled || unit

package mocks

import (
	"context"

	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/stretchr/testify/mock"
)

// StaticTime Round is used to remove monotonic clock from time.Now() to ensure that the time is compatible with equality checks
// var StaticTime = time.Now().UTC().Round(0)

type MockKfpApi2 struct {
	mock.Mock
}

func (m *MockKfpApi2) GetResourceReferences(
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

// func (mka *MockKfpApi2) ReturnResourceReferencesForRun() resource.References {
// 	mka.resourceReferences = resource.References{
// 		RunConfigurationName: common.RandomNamespacedName(),
// 		RunName:              common.RandomNamespacedName(),
// 		PipelineName:         common.RandomNamespacedName(),
// 		CreatedAt:            &StaticTime,
// 		FinishedAt:           &StaticTime,
// 	}
// 	mka.err = nil
//
// 	return mka.resourceReferences
// }
