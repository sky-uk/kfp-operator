//go:build unit
// +build unit

package main

import (
	"k8s.io/apimachinery/pkg/util/rand"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Update Unit Suite")
}

func randomString() string {
	return rand.String(5)
}
