//go:build decoupled || unit

package mocks

import (
	"context"

	"github.com/sky-uk/kfp-operator/pkg/common"
)

type MockMetadataStore struct {
	resultComponents []common.PipelineComponent
	err              error
}

func (mms *MockMetadataStore) GetArtifactsForRun(_ context.Context, _ string) ([]common.PipelineComponent, error) {
	return mms.resultComponents, mms.err

}

func (mms *MockMetadataStore) SetResultComponents(components []common.PipelineComponent) {
	mms.resultComponents = components
}

func (mms *MockMetadataStore) reset() {
	mms.err = nil
}

func (mms *MockMetadataStore) Reset() {
	mms.reset()
}

func (mms *MockMetadataStore) error(err error) {
	mms.err = err
}

func (mms *MockMetadataStore) Error(err error) {
	mms.error(err)
}
