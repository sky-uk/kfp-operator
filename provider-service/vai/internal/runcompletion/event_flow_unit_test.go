//go:build unit

package runcompletion

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VAI Event Flow Unit Suite")
}

func artifact() *aiplatformpb.Artifact {
	return &aiplatformpb.Artifact{
		SchemaTitle: "tfx.PushedModel", // Legacy
		DisplayName: "a-model",
		Uri:         "gs://some/where",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"pushed":             structpb.NewNumberValue(1),                 // Legacy
				"pushed_destination": structpb.NewStringValue("gs://some/where"), // Legacy
			},
		},
	}
}

var _ = Context("VaiEventingServer", func() {
	var (
		mockPipelineJobClient *mocks.MockPipelineJobClient
		eventingFlow          EventFlow
		inChan                chan StreamMessage[string]
		outChan               chan StreamMessage[*common.RunCompletionEventData]
		errChan               chan error
		handlerCall           chan any
		onCompHandlers        OnCompleteHandlers
	)

	BeforeEach(func() {
		mockPipelineJobClient = &mocks.MockPipelineJobClient{}
		inChan = make(chan StreamMessage[string])
		outChan = make(chan StreamMessage[*common.RunCompletionEventData])
		errChan = make(chan error)
		eventingFlow = EventFlow{
			ProviderConfig: config.VAIProviderConfig{
				Name: common.RandomString(),
			},
			PipelineJobClient: mockPipelineJobClient,
			context:           context.Background(),
			in:                inChan,
			out:               outChan,
			errorOut:          errChan,
		}
		handlerCall = make(chan any, 1)
		onCompHandlers = OnCompleteHandlers{
			OnSuccessHandler: func() {
				handlerCall <- "success_called"
			},
			OnUnrecoverableFailureHandler: func() {
				handlerCall <- "unrecoverable_failure_called"
			},
			OnRecoverableFailureHandler: func() {
				handlerCall <- "recoverable_failure_called"
			},
		}
	})

	DescribeTable("toRunCompletionEventData for job that has not completed", func(state aiplatformpb.PipelineState) {
		event, err := eventingFlow.toRunCompletionEventData(&aiplatformpb.PipelineJob{State: state}, common.RandomString())
		Expect(event).To(BeNil())
		Expect(err).To(HaveOccurred())
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_UNSPECIFIED),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_QUEUED),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_PENDING),
		Entry("Running", aiplatformpb.PipelineState_PIPELINE_STATE_RUNNING),
		Entry("Cancelling", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLING),
		Entry("Paused", aiplatformpb.PipelineState_PIPELINE_STATE_PAUSED),
	)

	DescribeTable("toRunCompletionEventData for job that has completed", func(pipelineState aiplatformpb.PipelineState, status common.RunCompletionStatus) {
		runConfigurationName := common.RandomNamespacedName()
		pipelineName := common.RandomNamespacedName()
		pipelineRunName := common.RandomNamespacedName()

		Expect(eventingFlow.toRunCompletionEventData(&aiplatformpb.PipelineJob{
			Name: pipelineRunName.Name,
			Labels: map[string]string{
				label.RunConfigurationName:      runConfigurationName.Name,
				label.RunConfigurationNamespace: runConfigurationName.Namespace,
				label.PipelineName:              pipelineName.Name,
				label.PipelineNamespace:         pipelineName.Namespace,
				label.RunName:                   pipelineRunName.Name,
				label.RunNamespace:              pipelineRunName.Namespace,
			},
			State: pipelineState,
			JobDetail: &aiplatformpb.PipelineJobDetail{
				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
					{
						TaskName: "my-task-name",
						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
							"a-model": {
								Artifacts: []*aiplatformpb.Artifact{
									artifact(),
								},
							},
						},
					},
				},
			},
		}, pipelineRunName.Name)).To(Equal(&common.RunCompletionEventData{
			RunConfigurationName: runConfigurationName.NonEmptyPtr(),
			PipelineName:         pipelineName,
			RunName:              pipelineRunName.NonEmptyPtr(),
			RunId:                pipelineRunName.Name,
			Status:               status,
			ServingModelArtifacts: []common.Artifact{
				{
					Name:     "a-model",
					Location: "gs://some/where",
				},
			},
			Provider: eventingFlow.ProviderConfig.Name,
			PipelineComponents: []common.PipelineComponent{
				{
					Name: "my-task-name",
					ComponentArtifacts: []common.ComponentArtifact{
						{
							Name: "a-model",
							Artifacts: []common.ComponentArtifactInstance{
								{
									Uri: "gs://some/where",
									Metadata: map[string]interface{}{
										"pushed_destination": "gs://some/where",
										"pushed":             float64(1),
									},
								},
							},
						},
					},
				},
			},
		}))
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED, common.RunCompletionStatuses.Succeeded),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, common.RunCompletionStatuses.Failed),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED, common.RunCompletionStatuses.Failed),
	)

	Describe("artifactsFilterData", func() {
		When("The job is missing the component", func() {
			It("Produces no artifacts", func() {
				Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{},
					},
				})).To(BeEmpty())
			})
		})

		When("The job has a component but no output", func() {
			It("Produces empty pipeline component", func() {
				componentName := common.RandomString()

				Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
							},
						},
					},
				})).To(Equal([]common.PipelineComponent{
					{
						Name:               componentName,
						ComponentArtifacts: []common.ComponentArtifact{},
					},
				}))
			})
		})

		When("The job has a component and output but no artifacts", func() {
			It("Produces empty pipeline component", func() {
				componentName := common.RandomString()

				Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: nil,
									},
								},
							},
						},
					},
				})).To(Equal([]common.PipelineComponent{
					{
						Name: componentName,
						ComponentArtifacts: []common.ComponentArtifact{
							{
								Name:      "a-model",
								Artifacts: []common.ComponentArtifactInstance{},
							},
						},
					},
				}))
			})
		})
	})

	Describe("Legacy: modelServingArtifactsForJob", func() {
		When("The job has an output with an artifact that doesn't match the SchemaTitle", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				incorrectArtifact.SchemaTitle = "a.Type"

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				})).To(Equal([]common.Artifact{}))
			})
		})

		When("The job has an output with an artifact that hasn't been pushed", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				incorrectArtifact.Metadata.Fields["pushed"] = structpb.NewNumberValue(0)

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				})).To(Equal([]common.Artifact{}))
			})
		})

		When("The job has an output with an artifact that isn't a float", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				incorrectArtifact.Metadata.Fields["pushed"] = structpb.NewStringValue("42")

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				})).To(Equal([]common.Artifact{}))
			})
		})

		When("The job has an output with an artifact that has no pushed property", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				delete(incorrectArtifact.Metadata.Fields, "pushed")

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				})).To(Equal([]common.Artifact{}))
			})
		})

		When("The job has an output with an artifact that has a pushed_destination that is not a string", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				incorrectArtifact.Metadata.Fields["pushed"] = structpb.NewNumberValue(42)

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				})).To(Equal([]common.Artifact{}))
			})
		})

		When("The job has an output with an artifact that has no pushed_destination property", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				delete(incorrectArtifact.Metadata.Fields, "pushed_destination")

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				})).To(Equal([]common.Artifact{}))
			})
		})

		When("The job has an output with several artifacts", func() {
			It("Produces several servingModelArtifacts", func() {
				firstArtifact := artifact()
				secondArtifact := artifact()
				secondArtifact.Metadata.Fields["pushed_destination"] = structpb.NewStringValue("gs://some/where/else")

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											firstArtifact,
											secondArtifact,
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(common.Artifact{Name: "a-model", Location: "gs://some/where"}, common.Artifact{Name: "a-model", Location: "gs://some/where/else"}))
			})
		})

		When("The job has several outputs with artifacts", func() {
			It("Produces several servingModelArtifacts", func() {
				firstArtifact := artifact()
				secondArtifact := artifact()
				secondArtifact.Metadata.Fields["pushed_destination"] = structpb.NewStringValue("gs://some/where/else")

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											firstArtifact,
										},
									},
									"another-model": {
										Artifacts: []*aiplatformpb.Artifact{
											secondArtifact,
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(common.Artifact{Name: "a-model", Location: "gs://some/where"}, common.Artifact{Name: "another-model", Location: "gs://some/where/else"}))
			})
		})

		When("The job has several tasks with artifacts", func() {
			It("Produces several servingModelArtifacts", func() {
				firstArtifact := artifact()
				secondArtifact := artifact()
				secondArtifact.Metadata.Fields["pushed_destination"] = structpb.NewStringValue("gs://some/where/else")

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											firstArtifact,
										},
									},
								},
							},
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"another-model": {
										Artifacts: []*aiplatformpb.Artifact{
											secondArtifact,
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(common.Artifact{Name: "a-model", Location: "gs://some/where"}, common.Artifact{Name: "another-model", Location: "gs://some/where/else"}))
			})
		})
	})

	Describe("runCompletionEventDataForRun", func() {
		When("GetPipelineJob errors", func() {
			It("returns no event", func() {
				runId := common.RandomString()
				expectedReq := aiplatformpb.GetPipelineJobRequest{
					Name: eventingFlow.ProviderConfig.PipelineJobName(runId),
				}
				expectedErr := errors.New("an error")
				mockPipelineJobClient.On(
					"GetPipelineJob",
					&expectedReq,
				).Return(nil, expectedErr)
				event, err := eventingFlow.runCompletionEventDataForRun(runId)
				Expect(event).To(BeNil())
				Expect(err).To(Equal(expectedErr))
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				runId := common.RandomString()
				expectedReq := aiplatformpb.GetPipelineJobRequest{
					Name: eventingFlow.ProviderConfig.PipelineJobName(runId),
				}
				mockPipelineJobClient.On(
					"GetPipelineJob",
					&expectedReq,
				).Return(
					&aiplatformpb.PipelineJob{
						State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
					},
					nil,
				)
				event, err := eventingFlow.runCompletionEventDataForRun(runId)
				Expect(event).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Start", func() {
		When("runCompletionEventDataForRun errors with NotFound", func() {
			It("acks the message and outputs to error sink", func() {
				runId := common.RandomString()
				expectedReq := aiplatformpb.GetPipelineJobRequest{
					Name: eventingFlow.ProviderConfig.PipelineJobName(runId),
				}
				expectedErr := status.New(codes.NotFound, "not found").Err()
				mockPipelineJobClient.On(
					"GetPipelineJob",
					&expectedReq,
				).Return(nil, expectedErr)
				eventingFlow.Start()

				eventingFlow.in <- StreamMessage[string]{Message: runId, OnCompleteHandlers: onCompHandlers}

				Eventually(handlerCall).Should(Receive(Equal("unrecoverable_failure_called")))
				Eventually(errChan).Should(Receive(Equal(expectedErr)))
			})
		})

		When("runCompletionEventDataForRun errors", func() {
			It("nacks the message and outputs to error sink", func() {
				runId := common.RandomString()
				expectedReq := aiplatformpb.GetPipelineJobRequest{
					Name: eventingFlow.ProviderConfig.PipelineJobName(runId),
				}
				expectedErr := errors.New("an error")
				mockPipelineJobClient.On(
					"GetPipelineJob",
					&expectedReq,
				).Return(nil, expectedErr)
				eventingFlow.Start()

				eventingFlow.in <- StreamMessage[string]{Message: runId, OnCompleteHandlers: onCompHandlers}

				Eventually(handlerCall).Should(Receive(Equal("recoverable_failure_called")))
				Eventually(errChan).Should(Receive(Equal(expectedErr)))
			})
		})

		When("runCompletionEventDataForRun succeeds", func() {
			It("sends the message to the out channel", func() {
				runId := common.RandomString()
				expectedReq := aiplatformpb.GetPipelineJobRequest{
					Name: eventingFlow.ProviderConfig.PipelineJobName(runId),
				}
				mockPipelineJobClient.On(
					"GetPipelineJob",
					&expectedReq,
				).Return(
					&aiplatformpb.PipelineJob{
						State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
					},
					nil,
				)

				eventingFlow.Start()
				eventingFlow.in <- StreamMessage[string]{Message: runId, OnCompleteHandlers: onCompHandlers}

				expectedRunCompletionEventData := &common.RunCompletionEventData{
					Status:                "succeeded",
					PipelineName:          common.NamespacedName{Name: "", Namespace: ""},
					RunConfigurationName:  nil,
					RunName:               nil,
					RunId:                 runId,
					ServingModelArtifacts: []common.Artifact{},
					PipelineComponents:    []common.PipelineComponent{},
					Provider:              eventingFlow.ProviderConfig.Name,
				}

				Eventually(outChan).Should(Receive(WithTransform(func(msg StreamMessage[*common.RunCompletionEventData]) interface{} {
					return msg.Message
				}, BeEquivalentTo(expectedRunCompletionEventData))))
			})
		})
	})
})
