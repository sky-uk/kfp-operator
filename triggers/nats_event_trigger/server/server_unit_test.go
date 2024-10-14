//go:build unit

package nats_event_trigger

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PublishFunc func(data []byte) error

func (pf PublishFunc) Publish(data []byte) error {
	return pf(data)
}

var _ = Context("ProcessEventFeed", func() {
	ctx := logr.NewContext(context.Background(), logr.Discard())

	When("nats returns an error", func() {
		It("returns Internal Error", func() {
			stubPublisher := struct {
				PublisherHandler
			}{
				PublishFunc(func(data []byte) error {
					return errors.New("an error")
				}),
			}

			stubServer := Server{
				UnimplementedNATSEventTriggerServer: pb.UnimplementedNATSEventTriggerServer{},
				Publisher:                           stubPublisher,
			}

			_, err := stubServer.ProcessEventFeed(ctx, &pb.RunCompletionEvent{})
			Expect(err).To(Equal(status.Error(codes.Internal, "failed to publish event")))
		})
	})

	When("successfully publishing to nats", func() {
		It("returns empty", func() {
			stubPublisher := struct {
				PublisherHandler
			}{
				PublishFunc(func(data []byte) error {
					return nil
				}),
			}

			stubServer := Server{
				UnimplementedNATSEventTriggerServer: pb.UnimplementedNATSEventTriggerServer{},
				Publisher:                           stubPublisher,
			}

			result, err := stubServer.ProcessEventFeed(ctx, &pb.RunCompletionEvent{})
			Expect(result).To(Equal(&emptypb.Empty{}))
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
