//go:build decoupled || unit
// +build decoupled unit

package kfp

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type MockMetadataStore struct {
	results []common.Artifact
	err     error
}

func (mms *MockMetadataStore) GetArtifacts(_ context.Context, _ string, _ []pipelinesv1.Artifact) ([]common.Artifact, error) {
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
