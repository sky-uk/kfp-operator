//go:build decoupled || unit

package kfp

import (
	"context"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type MockMetadataStore struct {
	servingArtifacts []common.Artifact
	artifacts        []common.Artifact
	err              error
}

func (mms *MockMetadataStore) GetArtifacts(_ context.Context, _ string, _ []pipelinesv1.OutputArtifact) ([]common.Artifact, error) {
	return mms.artifacts, mms.err
}

func (mms *MockMetadataStore) GetServingModelArtifact(_ context.Context, _ string) ([]common.Artifact, error) {
	return mms.servingArtifacts, mms.err
}

func (mms *MockMetadataStore) reset() {
	mms.servingArtifacts = nil
	mms.artifacts = nil
	mms.err = nil
}

func (mms *MockMetadataStore) setAndReturnServingArtifact() []common.Artifact {
	mms.servingArtifacts = []common.Artifact{common.RandomArtifact()}
	mms.err = nil

	return mms.servingArtifacts
}

func (mms *MockMetadataStore) setAndReturnArtifacts() []common.Artifact {
	mms.artifacts = []common.Artifact{common.RandomArtifact()}
	mms.err = nil

	return mms.artifacts
}

func (mms *MockMetadataStore) error(err error) {
	mms.servingArtifacts = nil
	mms.artifacts = nil
	mms.err = err
}
