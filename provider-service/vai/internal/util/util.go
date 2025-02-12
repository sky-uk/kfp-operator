package util

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/common"
)

func PipelineStorageObject(
	pipelineName common.NamespacedName,
	pipelineVersion string,
) (string, error) {
	namespaceName, err := pipelineName.String()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", namespaceName, pipelineVersion), nil
}

func PipelineUri(
	pipelineName common.NamespacedName,
	pipelineVersion string,
	bucket string,
) (string, error) {
	pipelineUri, err := PipelineStorageObject(pipelineName, pipelineVersion)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("gs://%s/%s", bucket, pipelineUri), nil
}
