//go:build unit

package test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	name               = "testPipeline"
	providerName       = "providerName"
	image              = "testImage"
	tfxComponentsKey   = "tfxComponentKey"
	tfxComponentsValue = "tfxComponentValue"
	beamArgsKey        = "beamArgsKey"
	beamArgsValue      = "beamArgsValue"
	envVarKey          = "envVarKey"
	envVarValue        = "envVarValue"
)

var _ = Context("Pipeline Hash", func() {

	var (
		beta1Pipeline = v1beta1.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: name,
			},
			Spec: v1beta1.PipelineSpec{
				Provider: common.NamespacedName{
					Name:      providerName,
					Namespace: "providerName",
				},
				Image: image,
				Env: []apis.NamedValue{
					{Name: envVarKey, Value: envVarValue},
				},
				Framework: v1beta1.PipelineFramework{
					Name: v1beta1.DefaultFallbackFramework,
					Parameters: map[string]*apiextensionsv1.JSON{
						"components": {
							Raw: []byte(fmt.Sprintf(`{"%s":"%s"}`, tfxComponentsKey, tfxComponentsValue)),
						},
						"beamArgs": {
							Raw: []byte(fmt.Sprintf(`{"%s":"%s"}`, beamArgsKey, beamArgsValue)),
						},
					},
				},
			},
		}

		alpha6Pipeline = v1alpha6.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: name,
			},
			Spec: v1alpha6.PipelineSpec{
				Provider: providerName,
				Image:    image,
				Env: []apis.NamedValue{
					{Name: envVarKey, Value: envVarValue},
				},
				TfxComponents: tfxComponentsValue,
				BeamArgs: []apis.NamedValue{
					{Name: beamArgsKey, Value: beamArgsValue},
				},
			},
		}

		alpha5Pipeline = v1alpha5.Pipeline{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: name,
			},
			Spec: v1alpha5.PipelineSpec{
				Image: image,
				Env: []apis.NamedValue{
					{Name: envVarKey, Value: envVarValue},
				},
				TfxComponents: tfxComponentsValue,
				BeamArgs: []apis.NamedValue{
					{Name: beamArgsKey, Value: beamArgsValue},
				},
			},
		}
	)

	var _ = Describe("ComputeVersion", func() {
		Specify("Should compute the same version for identical pipelines in different api specs", func() {
			beta1PipelineVersion := beta1Pipeline.ComputeVersion()
			alpha5pipelineVersion := alpha5Pipeline.ComputeVersion()
			alpha6PipelineVersion := alpha6Pipeline.ComputeVersion()

			Expect(alpha5pipelineVersion).To(Equal(beta1PipelineVersion))
			Expect(alpha6PipelineVersion).To(Equal(beta1PipelineVersion))
		})

		Specify("Should compute the same version for identical pipelines when beamArgs aren't set in different api specs", func() {
			delete(beta1Pipeline.Spec.Framework.Parameters, "beamArgs")
			beta1PipelineVersion := beta1Pipeline.ComputeVersion()

			alpha5Pipeline.Spec.BeamArgs = nil
			alpha5pipelineVersion := alpha5Pipeline.ComputeVersion()

			alpha6Pipeline.Spec.BeamArgs = nil
			alpha6PipelineVersion := alpha6Pipeline.ComputeVersion()

			Expect(alpha5pipelineVersion).To(Equal(beta1PipelineVersion))
			Expect(alpha6PipelineVersion).To(Equal(beta1PipelineVersion))
		})

		Specify("Should compute the same version for identical pipelines when components aren't set in different api specs", func() {
			delete(beta1Pipeline.Spec.Framework.Parameters, "components")
			beta1PipelineVersion := beta1Pipeline.ComputeVersion()

			alpha5Pipeline.Spec.TfxComponents = ""
			alpha5pipelineVersion := alpha5Pipeline.ComputeVersion()

			alpha6Pipeline.Spec.TfxComponents = ""
			alpha6PipelineVersion := alpha6Pipeline.ComputeVersion()

			Expect(alpha5pipelineVersion).To(Equal(beta1PipelineVersion))
			Expect(alpha6PipelineVersion).To(Equal(beta1PipelineVersion))
		})

	})
})
