//go:build integration

package main

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pb "github.com/sky-uk/kfp-operator/triggers/nats_event_trigger/proto"
	"google.golang.org/grpc"

	"github.com/nats-io/nats.go"
)

func TestNATSEventTriggerInt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NATs Event Trigger Integration Test Suite")
}

var _ = Context("NATSEventTriggerService", func() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	// Create a client for the NATSEventTrigger service
	client := pb.NewNATSEventTriggerClient(conn)

	When("the NATSEventTriggerService is called with a valid request", func() {
		It("returns empty and NATSEventBus receives an event", func() {
			event := &pb.RunCompletionEvent{
				PipelineName:         "my-pipeline",
				Provider:             "my-provider",
				RunConfigurationName: "my-run-config",
				RunId:                "run12345",
				RunName:              "run-name-123",
				Status:               "Succeeded",
				ServingModelArtifacts: []*pb.ServingModelArtifact{
					{
						Location: "s3://my-bucket/model-1",
						Name:     "model-1",
					},
					{
						Location: "s3://my-bucket/model-2",
						Name:     "model-2",
					},
				},
			}

			// Set a timeout for the RPC call
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

			natsURL := nats.DefaultURL
			nc, err := nats.Connect(natsURL)

			// Subscribe to a subject
			sub, err := nc.SubscribeSync("events")

			Expect(err).ToNot(HaveOccurred())
			// Call ProcessEventFeed RPC
			result, err := client.ProcessEventFeed(ctx, event)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())

			// Wait for a message
			msg, err := sub.NextMsg(10 * time.Second)
			Expect(err).ToNot(HaveOccurred())

			latestMsg := &pb.RunCompletionEvent{}

			err = json.Unmarshal(msg.Data, latestMsg)
			Expect(err).ToNot(HaveOccurred())

			eventJson, err := json.Marshal(event)
			Expect(err).ToNot(HaveOccurred())

			expected := &pb.RunCompletionEvent{}

			err = json.Unmarshal(eventJson, expected)
			Expect(err).ToNot(HaveOccurred())

			Expect(latestMsg).To(Equal(expected))

			nc.Close()
			conn.Close()
			cancel()
		})
	})
})
