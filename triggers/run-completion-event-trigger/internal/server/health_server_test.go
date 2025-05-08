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

	Context("request service is liveness", func() {
		It("returns a SERVING response", func() {
			result, err := healthServer.Check(
				ctx,
				&grpc_health_v1.HealthCheckRequest{Service: "liveness"},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_SERVING))
		})
	})

	Context("request service is readiness", func() {
		When("HealthCheck dependency returns true", func() {
			It("returns a SERVING response", func() {
				mockHealthCheck.On("IsHealthy").Return(true)
				result, err := healthServer.Check(
					ctx, &grpc_health_v1.HealthCheckRequest{Service: "readiness"},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_SERVING))
			})
		})

		When("HealthCheck dependency returns false", func() {
			It("returns a NOT_SERVING response", func() {
				mockHealthCheck.On("IsHealthy").Return(false)
				result, err := healthServer.Check(
					ctx,
					&grpc_health_v1.HealthCheckRequest{Service: "readiness"},
				)

				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_NOT_SERVING))
			})
		})
	})

	Context("request service is not liveness or readiness", func() {
		It("returns SERVICE_UNKNOWN", func() {
			result, err := healthServer.Check(
				ctx,
				&grpc_health_v1.HealthCheckRequest{Service: "other"},
			)

			Expect(err).To(MatchError("Unexpected service name [other]. Expected 'liveness' or 'readiness'"))
			Expect(result.GetStatus()).To(Equal(grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN))

		})
	})
})
