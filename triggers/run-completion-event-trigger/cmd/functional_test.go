//go:build functional

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/publisher"
	"log"
	"testing"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/internal/converters"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"google.golang.org/grpc"

	"github.com/nats-io/nats.go"
)

func TestRunCompletionEventTriggerFunctional(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Completion Event Trigger Functional Test Suite")
}

var _ = Context("RunCompletionEventTriggerService", Ordered, func() {

	var grpcConn *grpc.ClientConn
	var grpcClient pb.RunCompletionEventTriggerClient

	var natsConn *nats.Conn
	var natsSubscription *nats.Subscription

	BeforeAll(
		func() {
			grpcConn, err := grpc.NewClient(
				"localhost:50051",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Fatalf("Failed to connect to gRPC server: %v", err)
			}
			grpcClient = pb.NewRunCompletionEventTriggerClient(grpcConn)

			natsConn, err = nats.Connect(nats.DefaultURL)
			if err != nil {
				log.Fatalf("Failed to connect to gRPC server: %v", err)
			}

			natsSubscription, err = natsConn.SubscribeSync("events")
			if err != nil {
				log.Fatalf("Failed to connect to gRPC server: %v", err)
			}
		})

	AfterAll(func() {
		if grpcConn != nil {
			if err := grpcConn.Close(); err != nil {
				log.Fatalf("Failed to close gRPC connection: %v", err)
			}
		}
		if natsConn != nil {
			natsConn.Close()
		}
	})

	When("the Run Completion Event Trigger Service is called with a valid request", func() {
		It("returns empty and NATS receives an event", func() {
			artifacts := []*pb.Artifact{
				{
					Location: "gs://my-bucket/model-1",
					Name:     "model-1",
				},
				{
					Location: "gs://my-bucket/model-2",
					Name:     "model-2",
				},
			}

			runCompletionEvent := &pb.RunCompletionEvent{
				PipelineName:          "some-pipeline-name",
				Provider:              "some-provider",
				RunConfigurationName:  "some-run-configuration-name",
				RunId:                 "some-run-id",
				RunName:               "some-run-name",
				Status:                pb.Status_SUCCEEDED,
				ServingModelArtifacts: artifacts,
				Artifacts:             artifacts,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			result, err := grpcClient.ProcessEventFeed(ctx, runCompletionEvent)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())

			latestRunCompletionEventFromNats, err := getLatestMessageFromNats(natsSubscription)
			Expect(err).ToNot(HaveOccurred())

			expectedRunCompletionEvent, err := converters.ProtoRunCompletionToCommon(runCompletionEvent)
			Expect(err).ToNot(HaveOccurred())

			Expect(latestRunCompletionEventFromNats).To(Equal(&expectedRunCompletionEvent))
		})
	})
})

func getLatestMessageFromNats(natsSubscription *nats.Subscription) (*common.RunCompletionEvent, error) {
	msg, err := natsSubscription.NextMsg(5 * time.Second)
	Expect(err).ToNot(HaveOccurred())

	latestRunCompletionEventWrapped := &publisher.DataWrapper{}
	if err = json.Unmarshal(msg.Data, latestRunCompletionEventWrapped); err != nil {
		return nil, err
	}
	fmt.Printf("failed: %v", err)
	return &latestRunCompletionEventWrapped.Data, nil
}
