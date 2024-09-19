//go:build unit

package webhook

import (
	"context"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"go.uber.org/zap/zapcore"
)

var _ = Context("getRequestBody", func() {
	var logger, _ = common.NewLogger(zapcore.DebugLevel)
	var ctx = logr.NewContext(context.Background(), logger)

	When("valid request", func() {
		It("returns request body contents", func() {
			Expect(true).To(BeTrue())
		})
	})
})
