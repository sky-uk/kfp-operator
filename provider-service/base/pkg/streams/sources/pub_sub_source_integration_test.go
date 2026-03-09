//go:build integration

package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	pubsub "cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/internal/log"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

var topicNames = []string{}
var subscriptionNames = []string{}

const (
	pubsubSubscriptionName = "some_subscription"
	pubsubProject          = "some_project"
	pubsubHost             = "localhost:8085"
	maxDeliveryAttempts    = 5
	retryTimeout           = time.Second
)

func TestSourcesIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sources Integration Suite")
}

var _ = Context("Pub sub source", Ordered, func() {
	logger, err := log.NewLogger(zapcore.InfoLevel)
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel := createContextWithLogger(logger)

	BeforeAll(func() {
		err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		client, err := pubsub.NewClient(ctx, pubsubProject)
		if err == nil {
			defer client.Close()
			for _, topicName := range topicNames {
				_ = client.TopicAdminClient.DeleteTopic(
					ctx,
					&pubsubpb.DeleteTopicRequest{Topic: topicName},
				)
			}
			for _, subName := range subscriptionNames {
				_ = client.SubscriptionAdminClient.DeleteSubscription(
					ctx, &pubsubpb.DeleteSubscriptionRequest{Subscription: subName},
				)
			}
		}
		cancel()
	})

	BeforeEach(func() {
		cancel()
		ctx, cancel = createContextWithLogger(logger)
	})

	Describe("subscribing to a topic", func() {
		When("connected with valid messages", func() {
			It("should stream the messages", func() {
				pipelineId := "valid_pipeline"

				client, topicName, _, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				defer client.Close()
				source := createPubSubSource(ctx, client, pipelineId)

				message := LogEntry{
					Resource: Resource{
						Labels: map[string]string{PipelineJobLabel: pipelineId},
					},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, client, topicName, jsonMessage)

				result := <-source.Out()
				Expect(result.Message).To(Equal(pipelineId))
			})
		})

		When("the project does not exist", func() {
			It("should return an error", func() {
				client, err := pubsub.NewClient(ctx, "nonexistent_project")
				if err == nil {
					defer client.Close()
				}
				source, err := NewPubSubSource(ctx, "nonexistent_project", pubsubSubscriptionName, client)
				Expect(err).To(HaveOccurred())
				Expect(source).To(BeNil())
			})
		})

		When("the subscription cannot be established", func() {
			It("should return an error", func() {
				nonExistentSubscription := "nonexistent_subscription"
				client, err := pubsub.NewClient(ctx, pubsubProject)
				Expect(err).ToNot(HaveOccurred())
				defer client.Close()
				_, err = NewPubSubSource(ctx, pubsubProject, nonExistentSubscription, client)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("acknowledgements", func() {
		When("a streamed message is completed successfully", func() {
			It("should be ack'd", func() {
				pipelineId := "valid_message_ack_check"

				client, topicName, _, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				defer client.Close()
				source := createPubSubSource(ctx, client, pipelineId)

				message := LogEntry{
					Resource: Resource{
						Labels: map[string]string{
							PipelineJobLabel: pipelineId,
						}},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, client, topicName, jsonMessage)

				msg := <-source.Out()
				msg.OnSuccess()
				Expect(msg.Message).To(Equal(pipelineId))

				time.Sleep(retryTimeout * 2)

				select {
				case _ = <-source.Out():
					Fail("second message received")
				default:
					Succeed()
				}
			})
		})

		When("a streamed message is valid but fails upstream with a unrecoverable error", func() {
			It("should be ack'd", func() {
				pipelineId := "valid_message_unrecoverable_error_ack_check"

				client, topicName, _, _ := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				defer client.Close()
				source := createPubSubSource(ctx, client, pipelineId)

				message := LogEntry{
					Resource: Resource{
						Labels: map[string]string{
							PipelineJobLabel: pipelineId,
						}},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				sendMessage(ctx, client, topicName, jsonMessage)

				msg := <-source.Out()
				msg.OnUnrecoverableFailureHandler()
				Expect(msg.Message).To(Equal(pipelineId))

				time.Sleep(retryTimeout)

				select {
				case _ = <-source.Out():
					Fail("second message received")
				default:
					Succeed()
				}
			})
		})

		When("a streamed message is valid but fails upstream with a recoverable error", func() {
			It("should be nack'd", func() {
				pipelineId := "valid_message_recoverable_error_nack_check"

				client, topicName, _, deadletterSubName := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				defer client.Close()
				source := createPubSubSource(ctx, client, pipelineId)

				message := LogEntry{
					Resource: Resource{
						Labels: map[string]string{
							PipelineJobLabel: pipelineId,
						}},
				}

				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				deadletterOut := make(chan *pubsub.Message, 1)
				deadletterSubscriber := client.Subscriber(deadletterSubName)
				go func() {
					_ = deadletterSubscriber.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
						Expect(message.Data).To(Equal(jsonMessage))
						deadletterOut <- message
						return
					})
				}()

				sendMessage(ctx, client, topicName, jsonMessage)

				counter := 0
				for i := 0; i <= maxDeliveryAttempts; i++ {
					select {
					case msg := <-source.Out():
						counter++
						msg.OnRecoverableFailure()
					case <-time.After(retryTimeout * (maxDeliveryAttempts + 1)):
						break
					}
				}

				Expect(counter).To(Equal(maxDeliveryAttempts))

				select {
				case msg := <-deadletterOut:
					Expect(msg.Data).To(Equal(jsonMessage))
				case <-time.After(10 * time.Second):
					Fail("dead letter message not received")
				}
			})
		})

		When("a streamed message is malformed", func() {
			It("should be nack'd", func() {
				pipelineId := "invalid_message_nack_check"

				client, topicName, _, deadletterSubName := createClientTopicSubscription(ctx, pipelineId, pipelineId)
				defer client.Close()
				_ = createPubSubSource(ctx, client, pipelineId)

				message := LogEntry{
					Resource: Resource{Labels: map[string]string{"": ""}},
				}
				jsonMessage, err := json.Marshal(message)
				Expect(err).NotTo(HaveOccurred())

				deadletterOut := make(chan *pubsub.Message, 1)
				deadletterSubscriber := client.Subscriber(deadletterSubName)
				go func() {
					_ = deadletterSubscriber.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
						deadletterOut <- message
						return
					})
				}()

				sendMessage(ctx, client, topicName, jsonMessage)

				select {
				case msg := <-deadletterOut:
					Expect(msg.Data).To(Equal(jsonMessage))
				case <-time.After(10 * time.Second):
					Fail("dead letter message not received")
				}
			})
		})
	})
})

func createContextWithLogger(logger logr.Logger) (ctx context.Context, cancel context.CancelFunc) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(10000)*time.Millisecond)
	ctxWithLogger := logr.NewContext(ctxWithTimeout, logger)
	return ctxWithLogger, cancel
}

func sendMessage(ctx context.Context, client *pubsub.Client, topicName string, data []byte) {
	publisher := client.Publisher(topicName)
	msg := &pubsub.Message{
		Data: data,
	}
	publisher.Publish(ctx, msg)
}

func createTopic(ctx context.Context, client *pubsub.Client, topicName string) string {
	fqTopicName := fmt.Sprintf("projects/%s/topics/%s", pubsubProject, topicName)

	_, err := client.TopicAdminClient.CreateTopic(
		ctx, &pubsubpb.Topic{Name: fqTopicName})

	if err != nil && status.Code(err) != codes.AlreadyExists {
		Expect(err).ToNot(HaveOccurred())
	}

	return fqTopicName
}

func createSubIfNotExists(
	ctx context.Context,
	client *pubsub.Client,
	subscriptionName string,
	topicName string,
	deadLetterTopicName string,
	maxDeliveryAttempts int32,
	retryTimeout time.Duration,
) string {
	fqSubName := fmt.Sprintf("projects/%s/subscriptions/%s", pubsubProject, subscriptionName)

	_, err := client.SubscriptionAdminClient.GetSubscription(ctx, &pubsubpb.GetSubscriptionRequest{
		Subscription: fqSubName,
	})

	if err != nil && status.Code(err) == codes.NotFound {
		subpb := &pubsubpb.Subscription{
			Name:  fqSubName,
			Topic: topicName,
		}

		if deadLetterTopicName != "" {
			subpb.DeadLetterPolicy = &pubsubpb.DeadLetterPolicy{
				DeadLetterTopic:     deadLetterTopicName,
				MaxDeliveryAttempts: maxDeliveryAttempts,
			}
		}

		if retryTimeout > 0 {
			subpb.RetryPolicy = &pubsubpb.RetryPolicy{
				MinimumBackoff: durationpb.New(retryTimeout),
				MaximumBackoff: durationpb.New(retryTimeout),
			}
		}

		_, err = client.SubscriptionAdminClient.CreateSubscription(ctx, subpb)
	}

	Expect(err).ToNot(HaveOccurred())
	return fqSubName
}

func createClientTopicSubscription(
	ctx context.Context,
	topicName string,
	subscriptionName string,
) (*pubsub.Client, string, string, string) {
	deadLetterTopicName := fmt.Sprintf("deadletter_topic_%s", topicName)

	client, err := pubsub.NewClient(ctx, pubsubProject)
	Expect(err).ToNot(HaveOccurred())

	topic := createTopic(ctx, client, topicName)
	deadLetterTopic := createTopic(ctx, client, deadLetterTopicName)

	subscription := createSubIfNotExists(
		ctx, client, subscriptionName, topic,
		deadLetterTopic, maxDeliveryAttempts, retryTimeout,
	)

	deadSubscription := createSubIfNotExists(
		ctx, client, subscriptionName+"-deadletter",
		deadLetterTopic, "", 0, 0,
	)

	topicNames = append(topicNames, topic, deadLetterTopic)
	subscriptionNames = append(subscriptionNames, subscription, deadSubscription)
	return client, topic, subscription, deadSubscription
}

func createPubSubSource(
	ctx context.Context,
	client *pubsub.Client,
	subscription string,
) *PubSubSource {
	source, err := NewPubSubSource(ctx, pubsubProject, subscription, client)
	Expect(err).ToNot(HaveOccurred())
	return source
}
