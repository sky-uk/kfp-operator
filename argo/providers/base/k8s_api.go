package base

import (
	"context"
	"errors"
	"fmt"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var RunGVR = pipelinesv1.GroupVersion.WithResource("runs")
var RunConfigurationGVR = pipelinesv1.GroupVersion.WithResource("runconfigurations")

func artifactNamePathAsString(unstructuredArtifact map[string]interface{}) (string, string, error) {
	nameStr, ok := unstructuredArtifact["name"].(string)
	if !ok {
		err := errors.New("name in artifact is malformed")
		return "", "", err
	}
	pathStr, ok := unstructuredArtifact["path"].(string)
	if !ok {
		err := errors.New("path in artifact is malformed")
		return "", "", err
	}

	return nameStr, pathStr, nil
}

func artifactsForFields(ctx context.Context, obj *unstructured.Unstructured, fields ...string) ([]pipelinesv1.OutputArtifact, error) {
	logger := common.LoggerFromContext(ctx)
	artifactsField, hasArtifacts, err := unstructured.NestedSlice(obj.Object, fields...)
	if err != nil || !hasArtifacts {
		logger.Error(err, "Failed to get artifacts out of unstructured")
		return nil, err
	}

	return pipelines.MapErr(artifactsField, func(fa interface{}) (pipelinesv1.OutputArtifact, error) {
		unstructuredArtifact, ok := fa.(map[string]interface{})
		if !ok {
			err = errors.New("artifacts malformed")
			logger.Error(err, "Failed to cast")
			return pipelinesv1.OutputArtifact{}, err
		}

		nameStr, pathStr, err := artifactNamePathAsString(unstructuredArtifact)
		if err != nil {
			logger.Error(err, "Failed to extract data from unstructured artifact")
			return pipelinesv1.OutputArtifact{}, err
		}
		path, err := pipelinesv1.ArtifactPathFromString(pathStr)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to process path %+v", pathStr))
			return pipelinesv1.OutputArtifact{}, err
		}

		return pipelinesv1.OutputArtifact{
			Name: nameStr,
			Path: path,
		}, nil
	})
}

func artifactsForUnstructured(ctx context.Context, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) ([]pipelinesv1.OutputArtifact, error) {
	if gvr == RunGVR {
		return artifactsForFields(ctx, obj, "spec", "artifacts")
	}

	if gvr == RunConfigurationGVR {
		return artifactsForFields(ctx, obj, "spec", "run", "artifacts")
	}

	return nil, errors.New("unhandled resource, only runs and runconfigurations expected")
}

func CreateK8sClient() (dynamic.Interface, error) {
	k8sConfig, err := common.K8sClientConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

type K8sGetResource interface {
	GetNamespaceResource(K8sClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface
}

type K8sGetResourceImpl struct{}

type K8sApi struct {
	K8sGetResource
	K8sClient dynamic.Interface
}

func NewK8sApi(K8sClient dynamic.Interface) K8sApi {
	return K8sApi{
		K8sClient:      K8sClient,
		K8sGetResource: K8sGetResourceImpl{},
	}
}

func (k8a K8sGetResourceImpl) GetNamespaceResource(K8sClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	return K8sClient.Resource(gvr).Namespace(namespace)
}

func (k8a K8sApi) GetRunArtifactDefinitions(ctx context.Context, namespacedName types.NamespacedName, gvr schema.GroupVersionResource) ([]pipelinesv1.OutputArtifact, error) {
	obj, err := k8a.GetNamespaceResource(k8a.K8sClient, gvr, namespacedName.Namespace).Get(ctx, namespacedName.Name, metav1.GetOptions{})

	if err != nil {
		common.LoggerFromContext(ctx).Error(err, fmt.Sprintf("Failed to retrieve resource %+v", gvr))
		return nil, err
	}

	return artifactsForUnstructured(ctx, obj, gvr)
}
