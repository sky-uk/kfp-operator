//go:build unit
// +build unit

package objecthasher

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestObjectHasher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Object Hasher Suite")
}
