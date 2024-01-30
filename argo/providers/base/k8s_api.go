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
	"reflect"
)

var RunGVR = pipelinesv1.GroupVersion.WithResource("runs")
var RunConfigurationGVR = pipelinesv1.GroupVersion.WithResource("runconfigurations")

type FlattenedArtifact map[string]interface{}

func artifactsForFields(obj *unstructured.Unstructured, fields ...string) ([]pipelinesv1.OutputArtifact, error) {
	println(fmt.Sprintf("UNSTRUCTURED=%+v", obj))
	println(fmt.Sprintf("FIELDS=%+v", fields))

	artifactsField, hasArtifacts, err := unstructured.NestedSlice(obj.Object, fields...)
	if err != nil || !hasArtifacts {
		println(fmt.Sprintf("Failed to get artifacts out of unstructured %s", err))
		return nil, err
	}
	println(fmt.Sprintf("artifactsField=%+v", artifactsField))

	return pipelines.MapErr(artifactsField, func(fa interface{}) (pipelinesv1.OutputArtifact, error) {
		fmt.Println(fa)
		fmt.Println(reflect.TypeOf(fa))
		fmt.Println(fa.(map[string]interface{}))
		flattenedArtifact, ok := fa.(map[string]interface{})
		if !ok {
			println("Failed to cast")
			return pipelinesv1.OutputArtifact{}, errors.New("artifacts malformed")
		}
		println(fmt.Sprintf("flattenedArtifacts=%+v", flattenedArtifact))
		path, err := pipelinesv1.ArtifactPathFromString(flattenedArtifact["path"].(string))
		if err != nil {
			println(fmt.Sprintf("Failed to extract from map %s", err))
			return pipelinesv1.OutputArtifact{}, err
		}

		return pipelinesv1.OutputArtifact{
			Name: flattenedArtifact["name"].(string),
			Path: path,
		}, nil
	})
}

func artifactsForUnstructured(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) ([]pipelinesv1.OutputArtifact, error) {
	if gvr == RunGVR {
		return artifactsForFields(obj, "spec", "artifacts")
	}

	if gvr == RunConfigurationGVR {
		return artifactsForFields(obj, "spec", "run", "artifacts")
	}

	return nil, errors.New("unknown resource")
}

func CreateK8sClient() (dynamic.Interface, error) {
	k8sConfig, err := common.K8sClientConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(k8sConfig)
}

type K8sApi struct {
	K8sClient dynamic.Interface
}

func (k8a K8sApi) GetRunArtifactDefinitions(ctx context.Context, namespacedName types.NamespacedName, gvr schema.GroupVersionResource) ([]pipelinesv1.OutputArtifact, error) {
	obj, err := k8a.K8sClient.Resource(gvr).Namespace(namespacedName.Namespace).Get(ctx, namespacedName.Name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return artifactsForUnstructured(obj, gvr)
}
