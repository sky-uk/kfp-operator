//go:build unit

package server

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/mocks"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var _ = Context("Check", func() {
	var (
		mockHealthCheck mocks.MockHealthCheck
		healthServer    HealthServer
		ctx             = context.Background()
	)

	BeforeEach(func() {
		mockHealthCheck = mocks.MockHealthCheck{}
		healthServer = HealthServer{HealthCheck: &mockHealthCheck}
	})

	When("HealthCheck dependency returns true", func() {
		It("returns a SERVING response", func() {
			mockHealthCheck.On("IsHealthy").Return(true)
			result, err := healthServer.Check(ctx, &grpc_health_v1.HealthCheckRequest{})

			Expect(err).ToNot(HaveOccurred())
			Expect(result.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_SERVING))
		})
	})

	When("HealthCheck dependency returns false", func() {
		It("returns a NOT_SERVING response", func() {
			mockHealthCheck.On("IsHealthy").Return(false)
			result, err := healthServer.Check(ctx, &grpc_health_v1.HealthCheckRequest{})

			Expect(err).ToNot(HaveOccurred())
			Expect(result.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_NOT_SERVING))
		})
	})
})
