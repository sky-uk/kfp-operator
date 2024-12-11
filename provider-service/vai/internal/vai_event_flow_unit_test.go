//go:build unit

package internal

import (
	"context"
	"errors"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/provider-service/base/pkg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

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
		mockCtrl              *gomock.Controller
		mockPipelineJobClient *MockPipelineJobClient
		eventingFlow          VaiEventFlow
		inChan                chan StreamMessage[string]
		outChan               chan StreamMessage[*common.RunCompletionEventData]
		errChan               chan error
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockPipelineJobClient = NewMockPipelineJobClient(mockCtrl)
		inChan = make(chan StreamMessage[string])
		outChan = make(chan StreamMessage[*common.RunCompletionEventData])
		errChan = make(chan error)
		eventingFlow = VaiEventFlow{
			ProviderConfig: VAIProviderConfig{
				Name: common.RandomString(),
			},
			PipelineJobClient: mockPipelineJobClient,
			context:           context.Background(),
			in:                inChan,
			out:               outChan,
			errorOut:          errChan,
		}

	})

	AfterEach(func() {
		mockCtrl.Finish()
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

	//flow
	DescribeTable("toRunCompletionEventData for job that has completed", func(pipelineState aiplatformpb.PipelineState, status common.RunCompletionStatus) {
		runConfigurationName := common.RandomNamespacedName()
		pipelineName := common.RandomNamespacedName()
		pipelineRunName := common.RandomNamespacedName()

		Expect(eventingFlow.toRunCompletionEventData(&aiplatformpb.PipelineJob{
			Name: pipelineRunName.Name,
			Labels: map[string]string{
				labels.RunConfigurationName:      runConfigurationName.Name,
				labels.RunConfigurationNamespace: runConfigurationName.Namespace,
				labels.PipelineName:              pipelineName.Name,
				labels.PipelineNamespace:         pipelineName.Namespace,
				labels.RunName:                   pipelineRunName.Name,
				labels.RunNamespace:              pipelineRunName.Namespace,
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

	//flow
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

	//flow
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

	//flow
	Describe("runCompletionEventDataForRun", func() {
		When("GetPipelineJob errors", func() {
			It("returns no event", func() {
				expectedErr := errors.New("an error")
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
				event, err := eventingFlow.runCompletionEventDataForRun(common.RandomString())
				Expect(event).To(BeNil())
				Expect(err).To(Equal(expectedErr))
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(&aiplatformpb.PipelineJob{
					State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
				}, nil)
				event, err := eventingFlow.runCompletionEventDataForRun(common.RandomString())
				Expect(event).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("start", func() {
		When("runCompletionEventDataForRun errors", func() {
			It("fails to find the pipeline job", func() {
				handlerCall := make(chan any, 1)
				onCompHandlers := OnCompleteHandlers{
					OnSuccessHandler: func() {
						handlerCall <- "success_called"
					},
					OnFailureHandler: func() {
						handlerCall <- "failure_called"
					},
				}
				expectedErr := status.New(codes.NotFound, "not found").Err()
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
				eventingFlow.Start()

				eventingFlow.in <- StreamMessage[string]{Message: "a-run-id", OnCompleteHandlers: onCompHandlers}

				Eventually(handlerCall).Should(Receive(Equal("success_called")))
				Eventually(errChan).Should(Receive(Equal(expectedErr)))
			})
		})
	})
})
