//go:build unit

package webhook

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProcessEventFeedFunc func(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error)

func (pef ProcessEventFeedFunc) ProcessEventFeed(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return pef(ctx, in, opts...)
}

var _ = Context("call", func() {
	var ctx = logr.NewContext(context.Background(), logr.Discard())

	When("called", func() {
		rce := RandomRunCompletionEventData().ToRunCompletionEvent()
		protoRce, _ := RunCompletionEventToProto(rce)

		It("return no error when server responds with no error", func() {
			stubRunCompletionEventTrigger := struct {
				pb.RunCompletionEventTriggerClient
			}{
				ProcessEventFeedFunc(func(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					Expect(in).To(Equal(protoRce))
					return &emptypb.Empty{}, nil
				}),
			}

			stub := GrpcNatsTrigger{
				Upstream:          config.Endpoint{},
				Client:            stubRunCompletionEventTrigger,
				ConnectionHandler: func() error { return nil },
			}

			err := stub.call(ctx, rce)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error when server responds with an error", func() {
			testError := errors.New("some error")
			stubRunCompletionEventTrigger := struct {
				pb.RunCompletionEventTriggerClient
			}{
				ProcessEventFeedFunc(func(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, testError
				}),
			}

			stub := GrpcNatsTrigger{
				Upstream:          config.Endpoint{},
				Client:            stubRunCompletionEventTrigger,
				ConnectionHandler: func() error { return nil },
			}

			err := stub.call(ctx, rce)
			Expect(err).To(Equal(testError))
		})
	})
})

var _ = Context("EventDataToPbRunCompletion", func() {
	When("converting event data to proto run completion event", func() {
		namespacedName := common.NamespacedName{
			Name:      "name",
			Namespace: "namespace",
		}

		rce := common.RunCompletionEvent{
			Status:               "some-status",
			PipelineName:         namespacedName,
			RunConfigurationName: &namespacedName,
			RunName:              &namespacedName,
			RunId:                "some-runid",
			ServingModelArtifacts: []common.Artifact{
				{
					Name:     "some-artifact",
					Location: "gs://some/where",
				},
			},
			Provider: "some-provider",
		}

		It("return no error when event data is converted to proto runcompletion event", func() {
			protoRce, err := RunCompletionEventToProto(rce)
			expectedResult := &pb.RunCompletionEvent{
				Status:               "some-status",
				PipelineName:         "namespace/name",
				RunConfigurationName: "namespace/name",
				RunName:              "namespace/name",
				RunId:                "some-runid",
				ServingModelArtifacts: []*pb.ServingModelArtifact{
					{
						Name:     "some-artifact",
						Location: "gs://some/where",
					},
				},
				Provider: "some-provider",
			}
			Expect(protoRce).To(Equal(expectedResult))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error when namespaced name fields cannot be marshalled", func() {
			invalidNamespacedName := common.NamespacedName{
				Namespace: "namespace",
			}
			rce := common.RunCompletionEvent{
				Status:               "some-status",
				PipelineName:         namespacedName,
				RunConfigurationName: &namespacedName,
				RunName:              &namespacedName,
				RunId:                "some-runid",
				ServingModelArtifacts: []common.Artifact{
					{
						Name:     "some-artifact",
						Location: "gs://some/where",
					},
				},
				Provider: "some-provider",
			}
			pipelineTest := rce
			pipelineTest.PipelineName = invalidNamespacedName

			runNameTest := rce
			runNameTest.RunName = &invalidNamespacedName

			runConfigurationNameTest := rce
			runConfigurationNameTest.RunConfigurationName = &invalidNamespacedName

			testTargets := []common.RunCompletionEvent{
				pipelineTest,
				runNameTest,
				runConfigurationNameTest,
			}

			for i := range testTargets {
				_, err := RunCompletionEventToProto(testTargets[i])
				Expect(err).To(HaveOccurred())
			}
		})
	})
})
