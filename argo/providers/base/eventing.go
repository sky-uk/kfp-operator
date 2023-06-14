package base

import pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"

var LegacyArtifactDefinition = pipelinesv1.OutputArtifact{Name: "pushed_model", Path: pipelinesv1.ArtifactPath{
	Locator: pipelinesv1.ArtifactLocator{
		Component: "Pusher",
		Artifact: "pushed_model",
	}, Filter: "pushed == 1"}}
