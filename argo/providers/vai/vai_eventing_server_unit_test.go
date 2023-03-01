package vai

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/eventing"
	aiplatformpb "google.golang.org/genproto/googleapis/cloud/aiplatform/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func artifact() *aiplatformpb.Artifact {
	return &aiplatformpb.Artifact{
		SchemaTitle: "tfx.PushedModel",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"pushed":             structpb.NewNumberValue(1),
				"pushed_destination": structpb.NewStringValue("gs://some/where"),
			},
		},
	}
}

var _ = Context("VaiEventingServer", func() {
	var (
		mockCtrl              *gomock.Controller
		mockPipelineJobClient *MockPipelineJobClient
		eventingServer        VaiEventingServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockPipelineJobClient = NewMockPipelineJobClient(mockCtrl)
		eventingServer = VaiEventingServer{
			PipelineJobClient: mockPipelineJobClient,
			Logger:            logr.Discard(),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	DescribeTable("toRunCompletionEvent for job that has not completed", func(state aiplatformpb.PipelineState) {
		Expect(toRunCompletionEvent(&aiplatformpb.PipelineJob{State: state}, common.RandomString())).To(BeNil())
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_UNSPECIFIED),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_QUEUED),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_PENDING),
		Entry("Running", aiplatformpb.PipelineState_PIPELINE_STATE_RUNNING),
		Entry("Cancelling", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLING),
		Entry("Paused", aiplatformpb.PipelineState_PIPELINE_STATE_PAUSED),
	)

	Describe("modelServingArtifactsForJob", func() {
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
				})).To(BeEmpty())
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
				})).To(BeEmpty())
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
				})).To(BeEmpty())
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
				})).To(BeEmpty())
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
				})).To(BeEmpty())
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
				})).To(BeEmpty())
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
				})).To(ConsistOf(eventing.ServingModelArtifact{Name: "a-model", Location: "gs://some/where"}, eventing.ServingModelArtifact{Name: "a-model", Location: "gs://some/where/else"}))
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
				})).To(ConsistOf(eventing.ServingModelArtifact{Name: "a-model", Location: "gs://some/where"}, eventing.ServingModelArtifact{Name: "another-model", Location: "gs://some/where/else"}))
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
				})).To(ConsistOf(eventing.ServingModelArtifact{Name: "a-model", Location: "gs://some/where"}, eventing.ServingModelArtifact{Name: "another-model", Location: "gs://some/where/else"}))
			})
		})
	})

	DescribeTable("toRunCompletionEvent for job that has completed", func(pipelineState aiplatformpb.PipelineState, status eventing.RunCompletionStatus) {
		runConfigurationName := common.RandomString()
		pipelineName := common.RandomString()
		pipelineRunName := common.RandomNamespacedName()

		Expect(toRunCompletionEvent(&aiplatformpb.PipelineJob{
			Name: pipelineRunName.Name,
			Labels: map[string]string{
				labels.RunConfiguration: runConfigurationName,
				labels.PipelineName:     pipelineName,
				labels.Namespace:        pipelineRunName.Namespace,
			},
			State: pipelineState,
			JobDetail: &aiplatformpb.PipelineJobDetail{
				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
					{
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
		}, pipelineRunName.Name)).To(Equal(&eventing.RunCompletionEvent{
			RunConfigurationName: runConfigurationName,
			PipelineName:         pipelineName,
			RunName:              pipelineRunName,
			Status:               status,
			ServingModelArtifacts: []eventing.ServingModelArtifact{
				{
					Name:     "a-model",
					Location: "gs://some/where",
				},
			},
		}))
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED, eventing.RunCompletionStatuses.Succeeded),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, eventing.RunCompletionStatuses.Failed),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED, eventing.RunCompletionStatuses.Failed),
	)

	Describe("runCompletionEventForRun", func() {
		When("GetPipelineJob errors", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("an error"))
				event := eventingServer.runCompletionEventForRun(context.Background(), common.RandomString())
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob return no result", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), common.RandomString())
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(&aiplatformpb.PipelineJob{
					State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
				}, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), common.RandomString())
				Expect(event).NotTo(BeNil())
			})
		})
	})
})
