//go:build unit

package test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	name               = "testPipeline"
	providerName       = "providerName"
	image              = "testImage"
	tfxComponentsValue = "tfxComponentValue"
	beamArgsKey        = "beamArgsKey"
	beamArgsValue      = "beamArgsValue"
	envVarKey          = "envVarKey"
	envVarValue        = "envVarValue"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cross version api test suite")
}

var _ = Context("Pipeline Hash", Ordered, func() {

	var (
		beta1Pipeline  v1beta1.Pipeline
		alpha5Pipeline v1alpha5.Pipeline
		alpha6Pipeline v1alpha6.Pipeline
	)

	BeforeEach(
		func() {
			tfxComponentsJson, err := json.Marshal(tfxComponentsValue)
			Expect(err).To(Not(HaveOccurred()))

			beamArgs := []apis.NamedValue{{Name: beamArgsKey, Value: beamArgsValue}}
			beamArgsJson, err := json.Marshal(beamArgs)
			Expect(err).To(Not(HaveOccurred()))

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
						Name: v1beta1.FallbackFramework,
						Parameters: map[string]*apiextensionsv1.JSON{
							"components": {
								Raw: tfxComponentsJson,
							},
							"beamArgs": {
								Raw: beamArgsJson,
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

		})

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
