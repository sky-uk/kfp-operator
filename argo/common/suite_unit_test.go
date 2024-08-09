//go:build unit

package common

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommonUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common unit Suite")
}

var _ = Context("Marshal NamespacedName", Serial, func() {
	name := RandomString()
	namespace := RandomString()
	namespacedName := NamespacedName{Namespace: namespace, Name: name}

	When("value is provided", func() {
		It("custom marshaller is called", func() {
			serialised, err := json.Marshal(namespacedName)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(serialised)).To(Equal(`"` + namespace + "/" + name + `"`))
			deserialised := NamespacedName{}
			Expect(json.Unmarshal(serialised, &deserialised)).To(Succeed())
			Expect(deserialised).To(Equal(namespacedName))
		})
	})

	When("pointer is provider", func() {
		It("custom marshaller is called", func() {
			serialised, err := json.Marshal(&namespacedName)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(serialised)).To(Equal(`"` + namespace + "/" + name + `"`))
			deserialised := NamespacedName{}
			Expect(json.Unmarshal(serialised, &deserialised)).To(Succeed())
			Expect(deserialised).To(Equal(namespacedName))
		})
	})
})

var _ = Context("NamespacedName.String", Serial, func() {
	name := RandomString()
	namespace := RandomString()

	When("all fields are provided", func() {
		It("serialises into a '/' separated string", func() {
			serialised, err := NamespacedName{Namespace: namespace, Name: name}.String()
			Expect(err).NotTo(HaveOccurred())
			Expect(serialised).To(Equal(namespace + "/" + name))
		})
	})

	When("only a name is provided", func() {
		It("serialises only the name", func() {
			serialised, err := NamespacedName{Name: name}.String()
			Expect(err).NotTo(HaveOccurred())
			Expect(serialised).To(Equal(name))
		})
	})

	When("only namespace provided", func() {
		It("errors", func() {
			_, err := NamespacedName{Namespace: namespace}.String()
			Expect(err).To(HaveOccurred())
		})
	})

	When("nothing is provided", func() {
		It("serialises into the empty string", func() {
			serialised, err := NamespacedName{}.String()
			Expect(err).NotTo(HaveOccurred())
			Expect(serialised).To(BeEmpty())
		})
	})
})

var _ = Context("NamespacedNameFromString", Serial, func() {
	name := RandomString()
	namespace := RandomString()

	When("single `/` with fields", func() {
		It("deserialises into NamespacedName", func() {
			deserialised, err := NamespacedNameFromString(namespace + "/" + name)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialised).To(Equal(NamespacedName{Namespace: namespace, Name: name}))
		})
	})

	When("single `/` without fields", func() {
		It("errors", func() {
			_, err := NamespacedNameFromString("/" + name)
			Expect(err).To(HaveOccurred())
			_, err = NamespacedNameFromString(namespace + "/")
			Expect(err).To(HaveOccurred())
		})
	})

	When("multiple `/`", func() {
		It("deserialises into NamespacedName", func() {
			_, err := NamespacedNameFromString(namespace + "/" + name + "/" + RandomString())
			Expect(err).To(HaveOccurred())
		})
	})

	When("no `/`", func() {
		It("deserialises into empty name only", func() {
			deserialised, err := NamespacedNameFromString(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialised).To(Equal(NamespacedName{Name: name}))
		})
	})

	When("empty string", func() {
		It("deserialises into empty NamespacedName", func() {
			deserialised, err := NamespacedNameFromString("")
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialised).To(Equal(NamespacedName{}))
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

var _ = Context("RunCompletionEvent.Validate", func() {
	validEvent := RunCompletionEvent{
		Status: "succeeded",
		PipelineName: NamespacedName{
			Name:      "Name",
			Namespace: "Namespace",
		},
		RunConfigurationName: &NamespacedName{
			Name:      "Name",
			Namespace: "Namespace",
		},
		RunName: &NamespacedName{
			Name:      "Name",
			Namespace: "Namespace",
		},
		RunId:    "RunId",
		Provider: "Provider",
	}

	It("returns an error when Status is missing", func() {
		invalidEvent := validEvent
		invalidEvent.Status = ""
		Expect(invalidEvent.Validate()).To(Not(Succeed()))
	})

	It("returns an error when PipelineName is missing", func() {
		invalidEvent := validEvent
		invalidEvent.PipelineName = NamespacedName{}
		Expect(invalidEvent.Validate()).To(Not(Succeed()))
	})

	It("returns an error when Provider is missing", func() {
		invalidEvent := validEvent
		invalidEvent.Provider = ""
		Expect(invalidEvent.Validate()).To(Not(Succeed()))
	})

	It("returns an error when RunId is missing", func() {
		invalidEvent := validEvent
		invalidEvent.RunId = ""
		Expect(invalidEvent.Validate()).To(Not(Succeed()))
	})

	It("returnes an error when both RunConfigurationName and RunName are missing", func() {
		invalidEvent := validEvent
		invalidEvent.RunConfigurationName = nil
		invalidEvent.RunName = nil
		Expect(invalidEvent.Validate()).To(Not(Succeed()))
	})

	It("is valid when RunConfigurationName is missing and RunName is provided", func() {
		invalidEvent := validEvent
		invalidEvent.RunConfigurationName = nil
		Expect(invalidEvent.Validate()).To(Succeed())
	})

	It("is valid when RunName is missing and RunConfigurationName is provided", func() {
		invalidEvent := validEvent
		invalidEvent.RunName = nil
		Expect(invalidEvent.Validate()).To(Succeed())
	})

	It("returns an error when RunName and RunConfigurationName are nil", func() {
		invalidEvent := validEvent
		invalidEvent.RunName = nil
		invalidEvent.RunConfigurationName = nil
		Expect(invalidEvent.Validate()).To(Not(Succeed()))
	})

	It("returns an error when RunName inner struct Name field is empty and RunConfigurationName is nil", func() {
		invalidEvent := validEvent
		invalidEvent.RunName = &NamespacedName{
			Name:      "",
			Namespace: "Namespace",
		}
		invalidEvent.RunConfigurationName = nil

		Expect(invalidEvent.Validate()).To(Not(Succeed()))
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
