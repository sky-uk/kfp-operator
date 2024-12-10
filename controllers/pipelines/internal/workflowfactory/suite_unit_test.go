//go:build unit

package workflowfactory

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkflowFactoryUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controllers Workflow Factory Unit Suite")
}
