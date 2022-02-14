//go:build decoupled || unit
// +build decoupled unit

package run_completion

import "context"

type MockMetadataStore struct {
	results []ServingModelArtifact
	err     error
}

func (mms *MockMetadataStore) GetServingModelArtifact(_ context.Context, _ string) ([]ServingModelArtifact, error) {
	return mms.results, mms.err
}

func (mms *MockMetadataStore) reset() {
	mms.results = nil
	mms.err = nil
}

func (mms *MockMetadataStore) returnArtifactForPipeline() []ServingModelArtifact {
	mms.results = []ServingModelArtifact{
		{
			randomString(),
			randomString(),
		},
	}
	mms.err = nil

	return mms.results
}

func (mms *MockMetadataStore) error(err error) {
	mms.results = nil
	mms.err = err
}
