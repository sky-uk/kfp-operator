//go:build decoupled || unit

package mocks

import (
	"context"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"time"
)

// StaticTime Round is used to remove monotonic clock from time.Now() to ensure that the time is compatible with equality checks
var StaticTime = time.Now().UTC().Round(0)

type MockKfpApi struct {
	resourceReferences resource.References
	err                error
}

func (mka *MockKfpApi) GetResourceReferences(_ context.Context, _ string) (resource.References, error) {
	return mka.resourceReferences, mka.err
}

func (mka *MockKfpApi) Reset() {
	mka.resourceReferences = resource.References{}
	mka.err = nil
}

func (mka *MockKfpApi) ReturnResourceReferencesForRun() resource.References {
	mka.resourceReferences = resource.References{
		RunConfigurationName: common.RandomNamespacedName(),
		RunName:              common.RandomNamespacedName(),
		PipelineName:         common.RandomNamespacedName(),
		CreatedAt:            &StaticTime,
		FinishedAt:           &StaticTime,
	}
	mka.err = nil

	return mka.resourceReferences
}

func (mka *MockKfpApi) Error(err error) {
	mka.resourceReferences = resource.References{}
	mka.err = err
}
