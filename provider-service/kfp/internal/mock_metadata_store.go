//go:build decoupled || unit

package internal

import (
	"context"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
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

func (mms *MockMetadataStore) error(err error) {
	mms.results = nil
	mms.err = err
}
