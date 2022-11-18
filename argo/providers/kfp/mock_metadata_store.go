//go:build decoupled || unit
// +build decoupled unit

package kfp

import (
	"context"
	. "github.com/sky-uk/kfp-operator/providers/base"
)

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
			RandomString(),
			RandomString(),
		},
	}
	mms.err = nil

	return mms.results
}

func (mms *MockMetadataStore) error(err error) {
	mms.results = nil
	mms.err = err
}
