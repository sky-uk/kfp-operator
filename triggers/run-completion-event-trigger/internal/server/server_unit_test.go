//go:build unit

package server

import (
	"context"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/publisher"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PublishFunc func(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error

func (pf PublishFunc) Publish(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
	return pf(ctx, runCompletionEvent)
}

func (pf PublishFunc) Name() string {
	return "test-publisher"
}

func (pf PublishFunc) IsHealthy() bool {
	return true
}

var _ = Context("ProcessEventFeed", func() {
	ctx := logr.NewContext(context.Background(), logr.Discard())

	When("publisher returns a marshalling error", func() {
		It("returns Invalid Argument Error", func() {
			stubPublisher := struct {
				publisher.PublisherHandler
			}{
				PublishFunc(func(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
					return &publisher.MarshallingError{Message: "test error"}
				}),
			}

			stubServer := Server{
				UnimplementedRunCompletionEventTriggerServer: pb.UnimplementedRunCompletionEventTriggerServer{},
				Publisher: stubPublisher,
			}

			_, err := stubServer.ProcessEventFeed(ctx, &pb.RunCompletionEvent{})
			Expect(err).To(Equal(status.Error(codes.InvalidArgument, "failed to marshal event")))
		})
	})

	When("publisher returns a connection error", func() {
		It("returns Internal Error", func() {
			stubPublisher := struct {
				publisher.PublisherHandler
			}{
				PublishFunc(func(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
					return &publisher.ConnectionError{Message: "test error"}
				}),
			}

			stubServer := Server{
				UnimplementedRunCompletionEventTriggerServer: pb.UnimplementedRunCompletionEventTriggerServer{},
				Publisher: stubPublisher,
			}

			_, err := stubServer.ProcessEventFeed(ctx, &pb.RunCompletionEvent{})
			Expect(err).To(Equal(status.Error(codes.Internal, "publisher request to upstream failed")))
		})
	})

	When("successfully publishing to nats", func() {
		It("returns empty", func() {
			stubPublisher := struct {
				publisher.PublisherHandler
			}{
				PublishFunc(func(ctx context.Context, runCompletionEvent common.RunCompletionEvent) error {
					return nil
				}),
			}

			stubServer := Server{
				UnimplementedRunCompletionEventTriggerServer: pb.UnimplementedRunCompletionEventTriggerServer{},
				Publisher: stubPublisher,
			}

			result, err := stubServer.ProcessEventFeed(ctx, &pb.RunCompletionEvent{})
			Expect(result).To(Equal(&emptypb.Empty{}))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
