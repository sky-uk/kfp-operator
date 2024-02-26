//go:build unit

package vai

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("VAI Config", func() {

	config := VAIProviderConfig{}

	DescribeTable("getMaxConcurrentRunCountOrDefault", func(setValue int64, expectDefault bool) {
		config.MaxConcurrentRunCount = setValue
		expectedRes := setValue
		if expectDefault {
			expectedRes = 10
		}

		Expect(config.getMaxConcurrentRunCountOrDefault()).To(Equal(expectedRes))

	},
		Entry("", int64(0), true),
		Entry("", -common.RandomInt64(), true),
		Entry("", common.RandomInt64()+1, false),
	)
})
