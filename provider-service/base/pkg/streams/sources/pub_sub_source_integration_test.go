//go:build integration

package sources

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
	"time"
)

var topics = []*pubsub.Topic{}
var subscriptions = []*pubsub.Subscription{}

const (
	pubsubSubscriptionName = "some_subscription"
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

func createContextWithLogger(logger logr.Logger) (ctx context.Context, cancel context.CancelFunc) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)
	ctxWithLogger := logr.NewContext(ctxWithTimeout, logger)
	return ctxWithLogger, cancel
}

var _ = Context("Pub sub source", Ordered, func() {
	logger, err := common.NewLogger(zapcore.InfoLevel)
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel := createContextWithLogger(logger)

	appCtx, appCancel := createContextWithLogger(logger)

	BeforeAll(func() {
		err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		for _, topic := range topics {
			_ = topic.Delete(ctx)
		}
		for _, sub := range subscriptions {
			_ = sub.Delete(ctx)
		}
		cancel()
		appCancel()
	})

	BeforeEach(func() {
		cancel()
		ctx, cancel = createContextWithLogger(logger)
		appCtx, appCancel = createContextWithLogger(logger)
	})

	Describe("subscribing to a topic", func() {
		When("connected with valid messages", func() {
			It("should stream the messages", func() {
				pipelineId := "valid_pipeline"

				topic, _, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
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

				topic, _, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
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

		When("a streamed message is valid but fails upstream", func() {
			It("should be nack'd", func() {
				pipelineId := "valid_message_nack_check"

				topic, _, deadletterSub := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				underTest := createPubSubSource(appCtx, pipelineId)
				message := LogEntry{
					Resource: Resource{
						Labels: map[string]string{
							PipelineJobLabel: pipelineId,
						}},
				}
				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				deadMessages := false
				go func() {
					_ = deadletterSub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
						Expect(message.Data).To(Equal(jsonMessage))
						deadMessages = true
						return
					})
				}()

				sendMessage(ctx, topic, pipelineId, jsonMessage)
				timeout := 2 * time.Second
				counter := 0
				for i := 0; i < 6; i++ {
					select {
					case msg := <-underTest.Out():
						msg.OnFailure()
						counter = counter + 1
					case <-time.After(timeout):
						break
					}
				}

				Expect(counter).To(Equal(5))
				Expect(deadMessages).To(BeTrue())
			})
		})

		When("a streamed message is malformed it", func() {
			It("should be nack'd", func() {
				pipelineId := "invalid_message_nack_check"

				topic, _, deadletterSub := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				_ = createPubSubSource(appCtx, pipelineId)
				message := LogEntry{
					Resource: Resource{Labels: map[string]string{"": ""}},
				}
				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				deadMessages := false
				go func() {
					_ = deadletterSub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
						deadMessages = true
						Expect(message.Data).To(Equal(jsonMessage))
						return
					})
				}()

				sendMessage(ctx, topic, pipelineId, jsonMessage)

				time.Sleep(2 * time.Second)
				appCancel()

				Expect(deadMessages).To(BeTrue())
			})
		})
	})
})

func createTopicIfNotExists(ctx context.Context, client *pubsub.Client, topicName string) *pubsub.Topic {
	topic := client.Topic(topicName)
	topicExists, err := topic.Exists(ctx)
	Expect(err).ToNot(HaveOccurred())
	if err != nil || !topicExists {
		topic, err = client.CreateTopic(ctx, topicName)
		Expect(err).ToNot(HaveOccurred())
	}
	return topic
}

func createSubIfNotExists(ctx context.Context, client *pubsub.Client, subscriptionName string, topic *pubsub.Topic, deadLetterPol *pubsub.DeadLetterPolicy) *pubsub.Subscription {
	subscription := client.Subscription(subscriptionName)
	subscriptionExists, err := subscription.Exists(ctx)

	if err != nil || !subscriptionExists {
		subscription, err = client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			DeadLetterPolicy: deadLetterPol,
			Topic:            topic,
		})
		Expect(err).ToNot(HaveOccurred())
	}
	return subscription
}

func createClientTopicSubscription(ctx context.Context, topicName string, subscriptionName string) (*pubsub.Topic, *pubsub.Subscription, *pubsub.Subscription) {
	deadLetterTopicName := fmt.Sprintf("deadletter_topic_%s", topicName)
	client, err := pubsub.NewClient(ctx, pubsubProject)
	Expect(err).ToNot(HaveOccurred())

	topic := createTopicIfNotExists(ctx, client, topicName)
	deadLetterTopic := createTopicIfNotExists(ctx, client, deadLetterTopicName)

	subscription := createSubIfNotExists(ctx, client, subscriptionName, topic, &pubsub.DeadLetterPolicy{
		DeadLetterTopic:     deadLetterTopic.String(),
		MaxDeliveryAttempts: 5,
	})

	deadSubscription := createSubIfNotExists(ctx, client, subscriptionName+"-deadletter", deadLetterTopic, nil)
	Expect(err).ToNot(HaveOccurred())

	topics = append(topics, topic, deadLetterTopic)
	subscriptions = append(subscriptions, subscription, deadSubscription)
	return topic, subscription, deadSubscription
}

func createPubSubSource(ctx context.Context, subscription string) *PubSubSource {
	source, err := NewPubSubSource(ctx, pubsubProject, subscription)
	Expect(err).ToNot(HaveOccurred())
	return source
}
