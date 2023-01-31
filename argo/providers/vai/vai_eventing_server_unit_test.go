package vai

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/providers/base"
	aiplatformpb "google.golang.org/genproto/googleapis/cloud/aiplatform/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func artifact(pushedDestination string) *aiplatformpb.Artifact {
	return &aiplatformpb.Artifact{
		SchemaTitle: "tfx.PushedModel",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"pushed":             structpb.NewNumberValue(1),
				"pushed_destination": structpb.NewStringValue(pushedDestination),
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
			Logger: logr.Discard(),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	DescribeTable("toRunCompletionEvent for job that has not completed", func(state aiplatformpb.PipelineState) {
		Expect(toRunCompletionEvent(&aiplatformpb.PipelineJob{State: state})).To(BeNil())
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											{
												SchemaTitle: "a.Type",
												Metadata: &structpb.Struct{
													Fields: map[string]*structpb.Value{
														"pushed":             structpb.NewNumberValue(1),
														"pushed_destination": structpb.NewStringValue("gs://some/where"),
													},
												},
											},
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											{
												SchemaTitle: "a.Type",
												Metadata: &structpb.Struct{
													Fields: map[string]*structpb.Value{
														"pushed":             structpb.NewNumberValue(0),
														"pushed_destination": structpb.NewStringValue("gs://some/where"),
													},
												},
											},
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											{
												SchemaTitle: "a.Type",
												Metadata: &structpb.Struct{
													Fields: map[string]*structpb.Value{
														"pushed":             structpb.NewStringValue("42"),
														"pushed_destination": structpb.NewStringValue("gs://some/where"),
													},
												},
											},
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											{
												SchemaTitle: "a.Type",
											},
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											{
												SchemaTitle: "a.Type",
												Metadata: &structpb.Struct{
													Fields: map[string]*structpb.Value{
														"pushed":             structpb.NewNumberValue(1),
														"pushed_destination": structpb.NewNumberValue(42),
													},
												},
											},
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											{
												SchemaTitle: "a.Type",
												Metadata: &structpb.Struct{
													Fields: map[string]*structpb.Value{
														"pushed":             structpb.NewNumberValue(1),
													},
												},
											},
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
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											artifact("gs://some/where"),
											artifact("gs://some/where/else"),
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(base.ServingModelArtifact{Name: "a-model", Location: "gs://some/where"}, base.ServingModelArtifact{Name: "a-model", Location: "gs://some/where/else"}))
			})
		})

		When("The job has several outputs with artifacts", func() {
			It("Produces several servingModelArtifacts", func() {
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											artifact("gs://some/where"),
										},
									},
									"another-model": {
										Artifacts: []*aiplatformpb.Artifact{
											artifact("gs://some/where/else"),
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(base.ServingModelArtifact{Name: "a-model", Location: "gs://some/where"}, base.ServingModelArtifact{Name: "another-model", Location: "gs://some/where/else"}))
			})
		})

		When("The job has several tasks with artifacts", func() {
			It("Produces several servingModelArtifacts", func() {
				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"a-model": {
										Artifacts: []*aiplatformpb.Artifact{
											artifact("gs://some/where"),
										},
									},
								},
							},
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"another-model": {
										Artifacts: []*aiplatformpb.Artifact{
											artifact("gs://some/where/else"),
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(base.ServingModelArtifact{Name: "a-model", Location: "gs://some/where"}, base.ServingModelArtifact{Name: "another-model", Location: "gs://some/where/else"}))
			})
		})
	})

	DescribeTable("toRunCompletionEvent for job that has completed", func(pipelineState aiplatformpb.PipelineState, status base.RunCompletionStatus) {
		runConfigurationName := base.RandomString()
		pipelineName := base.RandomString()

		Expect(toRunCompletionEvent(&aiplatformpb.PipelineJob{
			Labels: map[string]string{
				labels.RunConfiguration: runConfigurationName,
				labels.PipelineName:     pipelineName,
			},
			State: pipelineState,
			JobDetail: &aiplatformpb.PipelineJobDetail{
				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
					{
						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
							"a-model": {
								Artifacts: []*aiplatformpb.Artifact{
									artifact("gs://some/where"),
								},
							},
						},
					},
				},
			},
		})).To(Equal(&base.RunCompletionEvent{
			RunConfigurationName: runConfigurationName,
			PipelineName:         pipelineName,
			Status:               status,
			ServingModelArtifacts: []base.ServingModelArtifact{
				{
					Name:     "a-model",
					Location: "gs://some/where",
				},
			},
		}))
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED, base.Succeeded),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, base.Failed),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED, base.Failed),
	)

	Describe("runCompletionEventForRun", func() {
		When("GetPipelineJob errors", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("an error"))
				event := eventingServer.runCompletionEventForRun(context.Background(), base.RandomString())
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob return no result", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), base.RandomString())
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(&aiplatformpb.PipelineJob{
					State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
				}, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), base.RandomString())
				Expect(event).NotTo(BeNil())
			})
		})
	})
})
