package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultProviderNamespace string

var ResourceAnnotations = struct {
	ProviderNamespace string
}{
	ProviderNamespace: apis.Group + "/providerNamespace",
}

func convertProviderTo(
	provider string,
	remainderNamespace string,
) common.NamespacedName {
	var namespace = remainderNamespace
	if namespace == "" {
		namespace = DefaultProviderNamespace
	}

	return common.NamespacedName{
		Name:      provider,
		Namespace: namespace,
	}
}

func convertProviderFrom(
	provider common.NamespacedName,
	remainder *RunConversionRemainder,
) string {
	if provider.Namespace != DefaultProviderNamespace {
		remainder.ProviderNamespace = provider.Namespace
	}

	return provider.Name
}

func convertProviderFrom2(
	provider common.NamespacedName,
	remainder *ExperimentConversionRemainder,
) string {
	if provider.Namespace != DefaultProviderNamespace {
		remainder.ProviderNamespace = provider.Namespace
	}

	return provider.Name
}

func convertProviderFrom3(
	provider common.NamespacedName,
	remainder *RunScheduleConversionRemainder,
) string {
	if provider.Namespace != DefaultProviderNamespace {
		remainder.ProviderNamespace = provider.Namespace
	}

	return provider.Name
}

func convertProviderFrom4(
	provider common.NamespacedName,
	remainder *RunConfigurationConversionRemainder,
) string {
	if provider.Namespace != DefaultProviderNamespace {
		remainder.ProviderNamespace = provider.Namespace
	}

	return provider.Name
}

func convertProviderFrom5(
	provider common.NamespacedName,
	remainder *PipelineConversionRemainder,
) string {
	if provider.Namespace != DefaultProviderNamespace {
		remainder.ProviderNamespace = provider.Namespace
	}

	return provider.Name
}

func namespaceToProvider(provider string, namespace string) common.NamespacedName {
	if namespace == "" {
		namespace = DefaultProviderNamespace
	}

	return common.NamespacedName{
		Name:      provider,
		Namespace: namespace,
	}
}

func getProviderNamespaceAnnotation(resource v1.Object) string {
	if providerNamespace, ok := resource.GetAnnotations()[ResourceAnnotations.ProviderNamespace]; ok {
		return providerNamespace
	}
	return ""
}

func removeProviderNamespaceAnnotation(resource v1.Object) {
	delete(resource.GetAnnotations(), ResourceAnnotations.ProviderNamespace)
}
