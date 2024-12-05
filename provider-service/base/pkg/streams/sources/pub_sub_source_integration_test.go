//go:build integration

package sources

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"os"
	"testing"
	"time"
)

const (
	pubsubSubscriptionName = "some_subscription"
	pubsubTopicName        = "some_topic"
	pubsubProject          = "some_project"
	pubsubHost             = "localhost:8085"
)

var (
	pubsubClient       *pubsub.Client
	pubsubSubscription *pubsub.Subscription
	pubsubTopic        *pubsub.Topic
	pubsubSource       *PubSubSource
)

func TestSourcesIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sources Integration Suite")
}

type StubFlow struct {
	in     chan pkg.StreamMessage[string]
	errOut chan error
	out    chan pkg.StreamMessage[*common.RunCompletionEventData]
}

func sendMessage(ctx context.Context, id string, data []byte) {
	msg := &pubsub.Message{
		ID:   id,
		Data: data,
	}
	pubsubTopic.Publish(ctx, msg)
}

var _ = Context("Pub sub source", Ordered, func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)

	BeforeAll(func() {
		err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost)
		Expect(err).ToNot(HaveOccurred())

		pubsubClient, pubsubTopic, pubsubSubscription = createClientTopicSubscription(ctx)
		pubsubSource = createPubSubSource(ctx)
	})

	AfterAll(func() { cancel() })

	Describe("subscribing to a topic", func() {
		When("connected with valid messages", func() {
			It("should stream the messages", func() {
				pipelineId := "valid_pipeline"

				message := LogEntry{
					Resource: Resource{Labels: map[string]string{
						PipelineJobLabel: pipelineId,
					}},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, "sub_to_topic_valid", jsonMessage)

				result := <-pubsubSource.Out()
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

	//Describe("acknowledgements", func() {
	//	When("a streamed message is completed successfully", func() {
	//		It("should be ack'd on the topic", func() {
	//			pipelineId := "valid_message_ack_check"
	//			message := LogEntry{
	//				Resource: Resource{Labels: map[string]string{
	//					PipelineJobLabel: pipelineId,
	//				}},
	//			}
	//			jsonMessage, err := json.Marshal(message)
	//			Expect(err).NotTo(HaveOccurred())
	//
	//			sendMessage(ctx, "valid_message_ack_check", jsonMessage)
	//			result := <-pubsubSource.Out()
	//			Expect(result.Message).To(Equal(pipelineId))
	//
	//			When("a streamed message is malformed it should be nack'd on the topic", func() {
	//				It("should stream the messages", func() {
	//					Expect(1).To(Equal(1))
	//				})
	//			})
	//
	//			When("a streamed message is completed and failed it should be nack'd on the topic", func() {
	//				It("should return an error", func() {
	//					Expect(1).To(Equal(1))
	//				})
	//			})
	//		})
	//	})
	//})
})

func createClientTopicSubscription(ctx context.Context) (*pubsub.Client, *pubsub.Topic, *pubsub.Subscription) {
	client, err := pubsub.NewClient(ctx, pubsubProject)
	Expect(err).ToNot(HaveOccurred())

	topic := client.Topic(pubsubTopicName)
	topicExists, err := topic.Exists(ctx)
	if err != nil || !topicExists {
		topic, err = client.CreateTopic(ctx, pubsubTopicName)
		Expect(err).ToNot(HaveOccurred())
	}

	subscription := client.Subscription(pubsubSubscriptionName)
	subscriptionExists, err := subscription.Exists(ctx)
	if err != nil || !subscriptionExists {
		subscription, err = client.CreateSubscription(ctx, pubsubSubscriptionName, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		Expect(err).ToNot(HaveOccurred())
	}
	return client, topic, subscription
}

func createPubSubSource(ctx context.Context) *PubSubSource {
	source, err := NewPubSubSource(ctx, pubsubProject, pubsubSubscriptionName)
	Expect(err).ToNot(HaveOccurred())
	return source
}
