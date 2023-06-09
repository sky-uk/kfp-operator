package base

import pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"

var LegacyArtifactDefinition = pipelinesv1.Artifact{Name: "pushed_model", Path: pipelinesv1.ArtifactPathDefinition{
	Path: pipelinesv1.ArtifactPath{
		Component: "Pusher",
		Artifact: "pushed_model",
	}, Filter: "pushed == 1"}}
