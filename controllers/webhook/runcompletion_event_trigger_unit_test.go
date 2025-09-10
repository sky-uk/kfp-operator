//go:build unit

package webhook

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/timestamppb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/pkg/common"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProcessEventFeedFunc func(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error)

func (pef ProcessEventFeedFunc) ProcessEventFeed(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return pef(ctx, in, opts...)
}

var _ = Context("Handle", func() {
	When("called", func() {
		logger, _ := log.NewLogger(zapcore.DebugLevel)
		ctx := logr.NewContext(context.Background(), logger)

		rce := RandomRunCompletionEventData().ToRunCompletionEvent()
		protoRce, _ := RunCompletionEventToProto(rce)

		It("return no error when server responds with no error", func() {
			stubClient := struct {
				pb.RunCompletionEventTriggerClient
			}{
				ProcessEventFeedFunc(func(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					Expect(in).To(Equal(protoRce))
					return &emptypb.Empty{}, nil
				}),
			}

			trigger := RunCompletionEventTrigger{
				EndPoint:          config.Endpoint{},
				Client:            stubClient,
				ConnectionHandler: func() error { return nil },
			}

			err := trigger.Handle(ctx, rce)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error when server responds with an error", func() {
			testError := errors.New("some error")
			stubClient := struct {
				pb.RunCompletionEventTriggerClient
			}{
				ProcessEventFeedFunc(func(ctx context.Context, in *pb.RunCompletionEvent, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, testError
				}),
			}

			trigger := RunCompletionEventTrigger{
				EndPoint:          config.Endpoint{},
				Client:            stubClient,
				ConnectionHandler: func() error { return nil },
			}

			err := trigger.Handle(ctx, rce)
			Expect(err).To(Equal(&FatalError{testError.Error()}))
		})
	})
})

var _ = Context("RunCompletionEventToProto", func() {
	var (
		namespacedName = common.NamespacedName{
			Name:      "name",
			Namespace: "namespace",
		}
		artifacts = []common.Artifact{
			{
				Name:     "some-artifact",
				Location: "gs://some/where",
			},
		}
		timeNow = time.Now().UTC()
		rce     common.RunCompletionEvent
	)

	BeforeEach(func() {
		rce = common.RunCompletionEvent{
			Status:                common.RunCompletionStatuses.Succeeded,
			PipelineName:          namespacedName,
			RunConfigurationName:  &namespacedName,
			RunName:               &namespacedName,
			RunId:                 "some-runid",
			ServingModelArtifacts: artifacts,
			Artifacts:             artifacts,
			Provider:              "some-provider",
			RunStartTime:          &timeNow,
			RunEndTime:            &timeNow,
		}
	})

	When("event data is converted to proto run completion event", func() {
		It("returns the converted event and no error", func() {
			protoRce, err := RunCompletionEventToProto(rce)
			Expect(err).NotTo(HaveOccurred())

			expectedArtifacts := []*pb.Artifact{
				{
					Name:     "some-artifact",
					Location: "gs://some/where",
				},
			}
			expectedResult := &pb.RunCompletionEvent{
				Status:                pb.Status_SUCCEEDED,
				PipelineName:          "namespace/name",
				RunConfigurationName:  "namespace/name",
				RunName:               "namespace/name",
				RunId:                 "some-runid",
				ServingModelArtifacts: expectedArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              "some-provider",
				RunStartTime:          timestamppb.New(timeNow),
				RunEndTime:            timestamppb.New(timeNow),
			}
			Expect(protoRce).To(Equal(expectedResult))
		})
	})

	When("There are no artifacts", func() {
		It("returns empty slices", func() {
			rce.Artifacts = []common.Artifact{}
			rce.ServingModelArtifacts = []common.Artifact{}

			protoRce, err := RunCompletionEventToProto(rce)
			Expect(err).NotTo(HaveOccurred())

			emptySliceOfArtifacts := []*pb.Artifact{}

			Expect(protoRce.Artifacts).To(Equal(emptySliceOfArtifacts))
			Expect(protoRce.ServingModelArtifacts).To(Equal(emptySliceOfArtifacts))
		})
	})

	When("event has no RunName", func() {
		It("returns no error", func() {
			rce.RunName = nil
			protoRce, err := RunCompletionEventToProto(rce)
			Expect(err).NotTo(HaveOccurred())

			expectedArtifacts := []*pb.Artifact{
				{
					Name:     "some-artifact",
					Location: "gs://some/where",
				},
			}
			expectedResult := &pb.RunCompletionEvent{
				Status:                pb.Status_SUCCEEDED,
				PipelineName:          "namespace/name",
				RunConfigurationName:  "namespace/name",
				RunName:               "",
				RunId:                 "some-runid",
				ServingModelArtifacts: expectedArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              "some-provider",
				RunStartTime:          timestamppb.New(timeNow),
				RunEndTime:            timestamppb.New(timeNow),
			}
			Expect(protoRce).To(Equal(expectedResult))
		})
	})

	When("event has no RunConfigurationName", func() {
		It("returns no error", func() {
			rce.RunConfigurationName = nil
			protoRce, err := RunCompletionEventToProto(rce)
			Expect(err).NotTo(HaveOccurred())

			expectedArtifacts := []*pb.Artifact{
				{
					Name:     "some-artifact",
					Location: "gs://some/where",
				},
			}
			expectedResult := &pb.RunCompletionEvent{
				Status:                pb.Status_SUCCEEDED,
				PipelineName:          "namespace/name",
				RunConfigurationName:  "",
				RunName:               "namespace/name",
				RunId:                 "some-runid",
				ServingModelArtifacts: expectedArtifacts,
				Artifacts:             expectedArtifacts,
				Provider:              "some-provider",
				RunStartTime:          timestamppb.New(timeNow),
				RunEndTime:            timestamppb.New(timeNow),
			}
			Expect(protoRce).To(Equal(expectedResult))
		})
	})

	When("event has no RunConfigurationName and no RunName", func() {
		It("returns error", func() {
			rce.RunConfigurationName = nil
			rce.RunName = nil
			_, err := RunCompletionEventToProto(rce)
			Expect(err).To(HaveOccurred())
		})
	})

	When("NamespacedName fields cannot be marshalled", func() {
		It("returns error", func() {
			invalidNamespacedName := common.NamespacedName{
				Namespace: "namespace",
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
