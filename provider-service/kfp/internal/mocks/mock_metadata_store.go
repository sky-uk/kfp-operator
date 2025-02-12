//go:build decoupled || unit

package mocks

import (
	"context"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/common"
)

type MockMetadataStore struct {
	results []common.Artifact
	err     error
}

func (mms *MockMetadataStore) GetArtifacts(_ context.Context, _ string, _ []pipelinesv1.OutputArtifact) ([]common.Artifact, error) {
	return mms.results, mms.err
}

func (mms *MockMetadataStore) GetServingModelArtifact(_ context.Context, _ string) ([]common.Artifact, error) {
	return mms.results, mms.err
}

func (mms *MockMetadataStore) reset() {
	mms.results = nil
	mms.err = nil
}

func (mms *MockMetadataStore) Reset() {
	mms.reset()
}

// We expose these private methods to have access to the generated implementation in the test suite.
func (mms *MockMetadataStore) returnArtifactForPipeline() []common.Artifact {
	mms.results = []common.Artifact{
		{
			"artifact-name",
			"artifact-location",
		},
	}
	mms.err = nil

	return mms.results
}

func (mms *MockMetadataStore) ReturnArtifactForPipeline() []common.Artifact {
	return mms.returnArtifactForPipeline()
}

func (mms *MockMetadataStore) error(err error) {
	mms.results = nil
	mms.err = err
}

func (mms *MockMetadataStore) Error(err error) {
	mms.error(err)
}
