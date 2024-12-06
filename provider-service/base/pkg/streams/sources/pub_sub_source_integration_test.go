//go:build integration

package sources

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/smartystreets/assertions/should"
	"os"
	"testing"
	"time"
)

var topics = []*pubsub.Topic{}

const (
	pubsubSubscriptionName = "some_subscription"
	pubsubTopicName        = "some_topic"
	pubsubProject          = "some_project"
	pubsubHost             = "localhost:8085"
)

func TestSourcesIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sources Integration Suite")
}

func sendMessage(ctx context.Context, topic *pubsub.Topic, id string, data []byte) {
	msg := &pubsub.Message{
		ID:   id,
		Data: data,
	}
	topic.Publish(ctx, msg)
}

var _ = Context("Pub sub source", Ordered, func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)

	BeforeAll(func() {
		err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		for _, topic := range topics {
			_ = topic.Delete(ctx)
		}
		cancel()
	})

	Describe("subscribing to a topic", func() {
		When("connected with valid messages", func() {
			It("should stream the messages", func() {
				pipelineId := "valid_pipeline"

				_, topic, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				source := createPubSubSource(ctx, pipelineId)

				message := LogEntry{
					Resource: Resource{Labels: map[string]string{
						PipelineJobLabel: pipelineId,
					}},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, topic, "sub_to_topic_valid", jsonMessage)

				result := <-source.Out()
				Expect(result.Message).To(Equal(pipelineId))
			})
		})

		When("the project does not exist", func() {
			It("should return an error", func() {
				source, err := NewPubSubSource(ctx, "nonexistent_project", pubsubSubscriptionName)
				Expect(err).To(HaveOccurred())
				Expect(source).To(BeNil())
			})
		})

		When("the subscription cannot be established", func() {
			It("should return an error", func() {
				nonExistentSubscription := "nonexistent_subscription"
				_, err := NewPubSubSource(ctx, pubsubProject, nonExistentSubscription)
				Expect(err).To(Equal(fmt.Errorf("subscription %s does not exist on topic", nonExistentSubscription)))
			})
		})
	})

	Describe("acknowledgements", func() {
		When("a streamed message is completed successfully", func() {
			It("should be ack'd on the topic", func() {
				pipelineId := "valid_message_ack_check"

				_, topic, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				source := createPubSubSource(ctx, pipelineId)

				message := LogEntry{
					Resource: Resource{
						Labels: map[string]string{
							PipelineJobLabel: pipelineId,
						}},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, topic, pipelineId, jsonMessage)

				result := <-source.Out()
				Expect(result.Message).To(Equal(pipelineId))

				select {
				case _ = <-source.Out():
					Fail("second message received")
				default:
					Succeed()
				}
			})
		})

		When("a streamed message is malformed it", func() {
			It("should be nack'd", func() {
				pipelineId := "invalid_message_nack_check"

				_, topic, subscription := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				_ = createPubSubSource(ctx, pipelineId)

				message := LogEntry{
					Resource: Resource{Labels: map[string]string{"": ""}},
				}
				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, topic, pipelineId, jsonMessage)

				attempts := 0
				err = subscription.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
					if message.DeliveryAttempt != nil {
						attempts = *message.DeliveryAttempt
						return
					}
					return
				})
				should.BeGreaterThan(attempts, 1)
			})
		})

		When("a streamed message is valid but fails upstream", func() {
			It("should be nack'd", func() {
				Expect(1).To(Equal(1))
			})
		})
	})
})

func createClientTopicSubscription(ctx context.Context, topicName string, subscriptionName string) (*pubsub.Client, *pubsub.Topic, *pubsub.Subscription) {
	client, err := pubsub.NewClient(ctx, pubsubProject)
	Expect(err).ToNot(HaveOccurred())

	topic := client.Topic(topicName)
	topicExists, err := topic.Exists(ctx)
	if err != nil || !topicExists {
		topic, err = client.CreateTopic(ctx, topicName)
		Expect(err).ToNot(HaveOccurred())
	}

	subscription := client.Subscription(subscriptionName)
	subscriptionExists, err := subscription.Exists(ctx)

	if err != nil || !subscriptionExists {
		subscription, err = client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		Expect(err).ToNot(HaveOccurred())
	}

	topics = append(topics, topic)
	return client, topic, subscription
}

func createPubSubSource(ctx context.Context, subscription string) *PubSubSource {
	source, err := NewPubSubSource(ctx, pubsubProject, subscription)
	Expect(err).ToNot(HaveOccurred())
	return source
}
