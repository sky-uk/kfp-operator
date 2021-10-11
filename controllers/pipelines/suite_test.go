package pipelines

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPipelineController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controller Suite")
}
