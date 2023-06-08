//go:build decoupled || unit
// +build decoupled unit

package kfp

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type MockMetadataStore struct {
	results []common.Artifact
	err     error
}

func (mms *MockMetadataStore) GetServingModelArtifact(_ context.Context, _ string) ([]common.Artifact, error) {
	return mms.results, mms.err
}

func (mms *MockMetadataStore) reset() {
	mms.results = nil
	mms.err = nil
}

func (mms *MockMetadataStore) returnArtifactForPipeline() []common.Artifact {
	mms.results = []common.Artifact{
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
