//go:build unit

package common

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("RunCompletionEvent.Validate", func() {

	validate, _ = InitialiseValidation()

	namespaceName := NamespacedName{
		Name:      "Name",
		Namespace: "Namespace",
	}

	baseEvent := RunCompletionEvent{
		Status:               "succeeded",
		PipelineName:         namespaceName,
		RunConfigurationName: &namespaceName,
		RunName:              &namespaceName,
		RunId:                "RunId",
		Provider:             "Provider",
	}

	It("is valid when RunConfigurationName is missing and RunName is provided", func() {
		invalidEvent := baseEvent
		invalidEvent.RunConfigurationName = nil
		Expect(invalidEvent.Validate(validate)).To(Succeed())
	})

	It("is valid when RunName is missing and RunConfigurationName is provided", func() {
		invalidEvent := baseEvent
		invalidEvent.RunName = nil
		Expect(invalidEvent.Validate(validate)).To(Succeed())
	})

	It("returns an error when Status is missing", func() {
		invalidEvent := baseEvent
		invalidEvent.Status = ""
		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	It("returns an error when PipelineName is missing", func() {
		invalidEvent := baseEvent
		invalidEvent.PipelineName = NamespacedName{}
		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	It("returns an error when Provider is missing", func() {
		invalidEvent := baseEvent
		invalidEvent.Provider = ""
		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	It("returns an error when RunId is missing", func() {
		invalidEvent := baseEvent
		invalidEvent.RunId = ""
		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	It("returnes an error when both RunConfigurationName and RunName are present", func() {
		invalidEvent := baseEvent
		invalidEvent.RunName = &namespaceName
		invalidEvent.RunConfigurationName = &namespaceName
		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	It("returnes an error when both RunConfigurationName and RunName are missing", func() {
		invalidEvent := baseEvent
		invalidEvent.RunConfigurationName = nil
		invalidEvent.RunName = nil
		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	It("returns an error when RunName inner struct Name field is empty and RunConfigurationName is nil", func() {
		invalidEvent := baseEvent
		invalidEvent.RunName = &NamespacedName{
			Name:      "",
			Namespace: "Namespace",
		}
		invalidEvent.RunConfigurationName = nil

		Expect(invalidEvent.Validate(validate)).To(Not(Succeed()))
	})

	var _ = Context("validateNamespacedName", func() {
		namespacedName := NamespacedName{
			Name:      "Name",
			Namespace: "Namespace",
		}

		It("succeed when Name and Namespace are not empty", func() {
			Expect(validateNamespacedName(&namespacedName, "")).To(Succeed())
		})

		It("returns error when Name is empty", func() {
			invalidNamespacedName := namespacedName
			invalidNamespacedName.Name = ""
			Expect(validateNamespacedName(&invalidNamespacedName, "")).To(Not(Succeed()))
		})

		It("succeed when Namespace is empty", func() {
			invalidNamespacedName := namespacedName
			invalidNamespacedName.Namespace = ""
			Expect(validateNamespacedName(&invalidNamespacedName, "")).To(Succeed())
		})

		It("returns error when Name and Namespace are empty", func() {
			invalidNamespacedName := namespacedName
			invalidNamespacedName.Name = ""
			invalidNamespacedName.Namespace = ""
			Expect(validateNamespacedName(&invalidNamespacedName, "")).To(Not(Succeed()))
		})

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
