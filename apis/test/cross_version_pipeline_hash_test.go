package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	beta1Pipeline = v1beta1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "name",
		},
		Spec: v1beta1.PipelineSpec{
			Provider: common.NamespacedName{
				Name: "name",
			},
			Image: "aa",
			Env: []apis.NamedValue{
				{Name: "a", Value: "b"},
			},
			Framework: v1beta1.PipelineFramework{
				Name: v1beta1.DefaultFallbackFramework,
				Parameters: map[string]*apiextensionsv1.JSON{
					"tfxComponents": {
						Raw: []byte(`{"key":"value"}`),
					},
					"beamArgs": {
						Raw: []byte(`{"key":"value2"}`),
					},
				},
			},
		},
	}

	alpha6Pipeline = v1alpha6.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "name",
		},
		Spec: v1alpha6.PipelineSpec{
			Provider: "name",
			Image:    "aa",
			Env: []apis.NamedValue{
				{Name: "a", Value: "b"},
			},
			TfxComponents: "value",
			BeamArgs: []apis.NamedValue{
				{Name: "key", Value: "value2"},
			},
		},
	}

	alpha5pipeline = v1alpha5.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "name",
		},
		Spec: v1alpha5.PipelineSpec{
			Image: "aa",
			Env: []apis.NamedValue{
				{Name: "a", Value: "b"},
			},
			TfxComponents: "value",
			BeamArgs: []apis.NamedValue{
				{Name: "key", Value: "value2"},
			},
		},
	}
)

var _ = Context("Pipeline Hash", func() {
	var _ = Describe("ComputeVersion", func() {
		Specify("Should compute the same version for identical pipelines in different api specs", func() {
			beta1PipelineVersion := beta1Pipeline.ComputeVersion()
			Expect(alpha5pipeline.ComputeVersion()).To(Equal(beta1PipelineVersion))
			Expect(alpha6Pipeline.ComputeVersion()).To(Equal(beta1PipelineVersion))
		})
	})
})
