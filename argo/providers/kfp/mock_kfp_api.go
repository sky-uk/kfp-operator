//go:build decoupled || unit
// +build decoupled unit

package kfp

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type MockKfpApi struct {
	resourceReferences ResourceReferences
	err                error
}

func (mka *MockKfpApi) GetResourceReferences(_ context.Context, _ string) (ResourceReferences, error) {
	return mka.resourceReferences, mka.err
}

func (mka *MockKfpApi) reset() {
	mka.resourceReferences = ResourceReferences{}
	mka.err = nil
}

func (mka *MockKfpApi) returnResourceReferencesForRun() ResourceReferences {
	mka.resourceReferences = ResourceReferences{
		RunConfigurationName: common.RandomNamespacedName(),
		RunName:              common.RandomNamespacedName(),
	}
	mka.err = nil

	return mka.resourceReferences
}

func (mka *MockKfpApi) error(err error) {
	mka.resourceReferences = ResourceReferences{}
	mka.err = err
}
