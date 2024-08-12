//go:build unit

package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCommonUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common unit Suite")
}
