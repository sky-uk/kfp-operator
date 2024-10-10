//go:build unit

package webhook

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"net/http"
	"regexp"
	"strconv"
)

func extractHostPort(url string) (string, int) {
	regex := regexp.MustCompile("(?:[^:]+)://([^:]+):([0-9]+)")
	matches := regex.FindStringSubmatch(url)
	port, _ := strconv.Atoi(matches[2])
	return matches[1], port
}

type stubServer struct {
	ExpectedResponse *pb.RunCompletionEvent
	StatusCode       *codes.Code
	pb.UnimplementedNATSEventTriggerServer
}

func (s stubServer) ProcessEventFeed(_ context.Context, in *pb.RunCompletionEvent) (*emptypb.Empty, error) {
	defer GinkgoRecover()

	println("I got a request")
	Expect(in).To(Equal(s.ExpectedResponse))

	return &emptypb.Empty{}, nil
}

func withGRPCServer(statusCode codes.Code, expectedResponse *pb.RunCompletionEvent, f func(upstream GrpcNatsTrigger)) func() {
	return func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "localhost", "4770"))

		if err != nil {
			panic(fmt.Errorf("failed to listen on port %s: %v", "4770", err))
		}

		s := grpc.NewServer()
		pb.RegisterNATSEventTriggerServer(s, &stubServer{ExpectedResponse: expectedResponse, StatusCode: &statusCode})
		if err := s.Serve(lis); err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}

		underTest := NewGrpcNatsTrigger(config.Endpoint{
			Host: "localhost",
			Port: 4770,
			Path: "",
		})

		f(underTest)
	}
}

var _ = Context("call", func() {
	var ctx = logr.NewContext(context.Background(), logr.Discard())

	When("called", func() {
		rce := common.RunCompletionEvent{}
		headers := http.Header{"hello": []string{"world", "goodbye"}}

		eventData := EventData{
			Header:             headers,
			RunCompletionEvent: rce,
		}

		It("return no error", withGRPCServer(codes.OK, nil, func(underTest GrpcNatsTrigger) {
			err := underTest.call(ctx, eventData)
			Expect(err).NotTo(HaveOccurred())
		}))
	})
})
