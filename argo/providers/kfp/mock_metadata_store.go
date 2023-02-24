//go:build decoupled || unit
// +build decoupled unit

package kfp

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/eventing"
)

type MockMetadataStore struct {
	results []eventing.ServingModelArtifact
	err     error
}

func (mms *MockMetadataStore) GetServingModelArtifact(_ context.Context, _ string) ([]eventing.ServingModelArtifact, error) {
	return mms.results, mms.err
}

func (mms *MockMetadataStore) reset() {
	mms.results = nil
	mms.err = nil
}

func (mms *MockMetadataStore) returnArtifactForPipeline() []eventing.ServingModelArtifact {
	mms.results = []eventing.ServingModelArtifact{
		{
			common.RandomString(),
			common.RandomString(),
		},
	}
	mms.err = nil

	return mms.results
}

func (mms *MockMetadataStore) error(err error) {
	mms.results = nil
	mms.err = err
}
