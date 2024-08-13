//go:build unit

package common

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("RunCompletionEvent.Validate", func() {

	validate := validator.New()
	validate.RegisterStructValidation(RunCompletionEventValidation, RunCompletionEvent{})

	namespaceName := NamespacedName{
		Name:      "Name",
		Namespace: "Namespace",
	}

	baseEvent := RunCompletionEvent{
		Status:               "succeeded",
		PipelineName:         namespaceName,
		RunConfigurationName: nil,
		RunName:              nil,
		RunId:                "RunId",
		Provider:             "Provider",
	}

	It("is valid when RunConfigurationName is missing and RunName is provided", func() {
		event := baseEvent
		event.RunName = &namespaceName
		Expect(validate.Struct(event)).To(Succeed())
	})

	It("is valid when RunName is missing and RunConfigurationName is provided", func() {
		event := baseEvent
		event.RunConfigurationName = &namespaceName
		Expect(validate.Struct(event)).To(Succeed())
	})

	It("returns an error when Status is missing", func() {
		event := baseEvent
		event.Status = ""
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when PipelineName is missing", func() {
		event := baseEvent
		event.PipelineName = NamespacedName{}
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when Provider is missing", func() {
		event := baseEvent
		event.Provider = ""
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when RunId is missing", func() {
		event := baseEvent
		event.RunId = ""
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when both RunConfigurationName and RunName are present", func() {
		event := baseEvent
		event.RunName = &namespaceName
		event.RunConfigurationName = &namespaceName
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when both RunConfigurationName and RunName are missing", func() {
		event := baseEvent
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when RunName inner struct Name field is empty and RunConfigurationName is nil", func() {
		event := baseEvent
		event.RunName = &NamespacedName{
			Name:      "",
			Namespace: "Namespace",
		}
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})

	It("returns an error when RunName inner struct Name field is empty and RunConfigurationName is nil", func() {
		event := baseEvent
		event.RunName = &NamespacedName{
			Name:      "",
			Namespace: "Namespace",
		}
		Expect(validate.Struct(event)).To(Not(Succeed()))
	})
})

var _ = Context("RunCompletionEvent.String", func() {
	artList := []Artifact{
		{
			Name:     "ArtifactName",
			Location: "ArtifactLocation",
		},
	}
	fixedEvent := RunCompletionEvent{
		PipelineName: NamespacedName{
			Name:      "PipelineNameName",
			Namespace: "PipelineNameNamespace",
		},
		RunName: &NamespacedName{
			Name:      "RunNameName",
			Namespace: "RunNameNamespace",
		},
		RunConfigurationName: &NamespacedName{
			Name:      "RunConfigurationNameName",
			Namespace: "RunConfigurationNameNamespace",
		},
		RunId:                 "RunId",
		ServingModelArtifacts: artList,
		Artifacts:             artList,
		Provider:              "Provider",
	}

	It("returns a string representation including all fields", func() {
		Expect(fmt.Sprintf("%+v", fixedEvent)).To(
			Equal(
				"{Status: PipelineName:{Name:PipelineNameName Namespace:PipelineNameNamespace} RunConfigurationName:&{Name:RunConfigurationNameName " +
					"Namespace:RunConfigurationNameNamespace} RunName:&{Name:RunNameName Namespace:RunNameNamespace} RunId:RunId " +
					"ServingModelArtifacts:[{Name:ArtifactName Location:ArtifactLocation}] " +
					"Artifacts:[{Name:ArtifactName Location:ArtifactLocation}] " +
					"Provider:Provider}",
			),
		)
	})

	It("returns a string representation handling nil RunConfigurationName", func() {
		missingRunConfigName := fixedEvent
		missingRunConfigName.RunConfigurationName = nil

		Expect(fmt.Sprintf("%+v", missingRunConfigName)).To(
			Equal(
				"{Status: PipelineName:{Name:PipelineNameName Namespace:PipelineNameNamespace} RunConfigurationName:<nil> " +
					"RunName:&{Name:RunNameName Namespace:RunNameNamespace} RunId:RunId " +
					"ServingModelArtifacts:[{Name:ArtifactName Location:ArtifactLocation}] " +
					"Artifacts:[{Name:ArtifactName Location:ArtifactLocation}] " +
					"Provider:Provider}",
			),
		)
	})

	It("returns a string representation handling nil RunName", func() {
		missingRunName := fixedEvent
		missingRunName.RunName = nil

		Expect(fmt.Sprintf("%+v", missingRunName)).To(
			Equal(
				"{Status: PipelineName:{Name:PipelineNameName Namespace:PipelineNameNamespace} RunConfigurationName:&{Name:RunConfigurationNameName " +
					"Namespace:RunConfigurationNameNamespace} RunName:<nil> RunId:RunId " +
					"ServingModelArtifacts:[{Name:ArtifactName Location:ArtifactLocation}] " +
					"Artifacts:[{Name:ArtifactName Location:ArtifactLocation}] " +
					"Provider:Provider}",
			),
		)
	})
})
