//go:build unit
// +build unit

package common

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
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
			Expect(string(serialised)).To(Equal(`"`+namespace+"/"+name+`"`))
			deserialised := NamespacedName{}
			Expect(json.Unmarshal(serialised, &deserialised)).To(Succeed())
			Expect(deserialised).To(Equal(namespacedName))
		})
	})

	When("pointer is provider", func() {
		It("custom marshaller is called", func() {
			serialised, err := json.Marshal(&namespacedName)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(serialised)).To(Equal(`"`+namespace+"/"+name+`"`))
			deserialised := NamespacedName{}
			Expect(json.Unmarshal(serialised, &deserialised)).To(Succeed())
			Expect(deserialised).To(Equal(namespacedName))
		})
	})
})

var _ = Context("NamespacedName.string", Serial, func() {
	name := RandomString()
	namespace := RandomString()

	When("all fields are provided", func() {
		It("serialises into a '/' separated string", func() {
			serialised, err := NamespacedName{Namespace: namespace, Name: name}.string()
			Expect(err).NotTo(HaveOccurred())
			Expect(serialised).To(Equal(namespace+"/"+name))
		})
	})

	When("only a name is provided", func() {
		It("serialises only the name", func() {
			serialised, err := NamespacedName{Name: name}.string()
			Expect(err).NotTo(HaveOccurred())
			Expect(serialised).To(Equal(name))
		})
	})

	When("only namespace provided", func() {
		It("errors", func() {
			_, err := NamespacedName{Namespace: namespace}.string()
			Expect(err).To(HaveOccurred())
		})
	})

	When("nothing is provided", func() {
		It("serialises into the empty string", func() {
			serialised, err := NamespacedName{}.string()
			Expect(err).NotTo(HaveOccurred())
			Expect(serialised).To(BeEmpty())
		})
	})
})

var _ = Context("namespacedNameFromString", Serial, func() {
	name := RandomString()
	namespace := RandomString()

	When("single `/` with fields", func() {
		It("deserialises into NamespacedName", func() {
			deserialised, err := namespacedNameFromString(namespace+"/"+name)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialised).To(Equal(NamespacedName{Namespace: namespace, Name: name}))
		})
	})

	When("single `/` without fields", func() {
		It("errors", func() {
			_, err := namespacedNameFromString("/"+name)
			Expect(err).To(HaveOccurred())
			_, err = namespacedNameFromString(namespace+"/")
			Expect(err).To(HaveOccurred())
		})
	})

	When("multiple `/`", func() {
		It("deserialises into NamespacedName", func() {
			_, err := namespacedNameFromString(namespace+"/"+name+"/"+RandomString())
			Expect(err).To(HaveOccurred())
		})
	})

	When("no `/`", func() {
		It("deserialises into empty name only", func() {
			deserialised, err := namespacedNameFromString(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialised).To(Equal(NamespacedName{Name: name}))
		})
	})

	When("empty string", func() {
		It("deserialises into empty NamespacedName", func() {
			deserialised, err := namespacedNameFromString("")
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialised).To(Equal(NamespacedName{}))
		})
	})
})
