//go:build decoupled || unit

package internal

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
		// RunConfigurationName: common.RandomNamespacedName(),
		// RunName:              common.RandomNamespacedName(),
		// PipelineName:         common.RandomNamespacedName(),
		RunConfigurationName: common.NamespacedName{
			Name: "rc-name",
			Namespace: "rc-namespace",
		},
		RunName:              common.NamespacedName{
			Name: "r-name",
			Namespace: "r-namespace",
		},
		PipelineName:         common.NamespacedName{
			Name: "p-name",
			Namespace: "p-namespace",
		},
	}
	mka.err = nil

	return mka.resourceReferences
}

func (mka *MockKfpApi) error(err error) {
	mka.resourceReferences = ResourceReferences{}
	mka.err = err
}
