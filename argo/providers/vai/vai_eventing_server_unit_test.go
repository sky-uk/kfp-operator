package vai

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	aiplatformpb "google.golang.org/genproto/googleapis/cloud/aiplatform/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func artifact() *aiplatformpb.Artifact {
	return &aiplatformpb.Artifact{
		DisplayName: "a-model",
		Uri: "gs://some/where",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"pushed": structpb.NewNumberValue(1),
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
		Expect(toRunCompletionEvent(&aiplatformpb.PipelineJob{State: state}, VAIRun{RunId: common.RandomString()})).To(BeNil())
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_UNSPECIFIED),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_QUEUED),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_PENDING),
		Entry("Running", aiplatformpb.PipelineState_PIPELINE_STATE_RUNNING),
		Entry("Cancelling", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLING),
		Entry("Paused", aiplatformpb.PipelineState_PIPELINE_STATE_PAUSED),
	)

	Describe("artifactsForJob", func() {
		When("The job is missing the component", func() {
			It("Produces no artifacts", func() {
				artifactDefs := []pipelinesv1.Artifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPathDefinition{
						Path: pipelinesv1.ArtifactPath{
							Component: common.RandomString(),
						},
					},
				}}

				Expect(artifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{},
					},
				}, artifactDefs)).To(BeEmpty())
			})
		})

		When("The job has a matching component that misses the output", func() {
			It("Produces no artifacts", func() {
				componentName := common.RandomString()
				artifactDefs := []pipelinesv1.Artifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPathDefinition{
						Path: pipelinesv1.ArtifactPath{
							Component: componentName,
							Artifact:  common.RandomString(),
						},
					},
				}}

				Expect(artifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
							},
						},
					},
				}, artifactDefs)).To(BeEmpty())
			})
		})

		When("The job has a matching component and a matching output but the artifact has no uri", func() {
			It("Produces the artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				incorrectArtifact := artifact()
				incorrectArtifact.Uri = ""

				artifactDefs := []pipelinesv1.Artifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPathDefinition{
						Path: pipelinesv1.ArtifactPath{
							Component: componentName,
							Artifact:  outputName,
						},
					},
				}}

				Expect(artifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									outputName: {
										Artifacts: []*aiplatformpb.Artifact{
											incorrectArtifact,
										},
									},
								},
							},
						},
					},
				}, artifactDefs)).To(BeEmpty())
			})
		})

		When("The job has a matching component and a matching output but no matching properties", func() {
			It("Produces the artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				incorrectArtifact := artifact()
				incorrectArtifact.Metadata.Fields = nil

				artifactDefs := []pipelinesv1.Artifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPathDefinition{
						Path: pipelinesv1.ArtifactPath{
							Component: componentName,
							Artifact:  outputName,
						},
						Filter: "a == b",
					},
				}}

				Expect(artifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									outputName: {
										Artifacts: []*aiplatformpb.Artifact{
											artifact(),
										},
									},
								},
							},
						},
					},
				}, artifactDefs)).To(BeEmpty())
			})
		})

		When("The job has a matching component and a matching component with several artifacts", func() {
			It("Produces the artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				artifactName := common.RandomString()
				artifactDefs := []pipelinesv1.Artifact{{
					Name: artifactName,
					Path: pipelinesv1.ArtifactPathDefinition{
						Path: pipelinesv1.ArtifactPath{
							Component: componentName,
							Artifact:  outputName,
						},
					},
				}}

				firstArtifact := artifact()
				secondArtifact := artifact()
				secondArtifact.Uri = "gs://some/where/else"
				secondArtifact.DisplayName = "another-artifact"

				Expect(artifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									outputName: {
										Artifacts: []*aiplatformpb.Artifact{
											firstArtifact,
											secondArtifact,
										},
									},
								},
							},
						},
					},
				}, artifactDefs)).To(ConsistOf(common.Artifact{Name: artifactName, Location: "gs://some/where"}, common.Artifact{Name: artifactName, Location: "gs://some/where/else"}))
			})
		})

		When("The job has a matching component and a matching component with several artifacts that have matching properties", func() {
			It("Produces the artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				artifactName := common.RandomString()
				artifactDefs := []pipelinesv1.Artifact{{
					Name: artifactName,
					Path: pipelinesv1.ArtifactPathDefinition{
						Path: pipelinesv1.ArtifactPath{
							Component: componentName,
							Artifact:  outputName,
						},
						Filter: "pushed == 1",
					},
				}}

				firstArtifact := artifact()
				secondArtifact := artifact()
				secondArtifact.Uri = "gs://some/where/else"
				secondArtifact.DisplayName = "another-artifact"

				Expect(artifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: componentName,
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									outputName: {
										Artifacts: []*aiplatformpb.Artifact{
											firstArtifact,
											secondArtifact,
										},
									},
								},
							},
						},
					},
				}, artifactDefs)).To(ConsistOf(common.Artifact{Name: artifactName, Location: "gs://some/where"}, common.Artifact{Name: artifactName, Location: "gs://some/where/else"}))
			})
		})
	})

	Describe("Legacy: modelServingArtifactsForJob", func() {
		When("The job has an output with an artifact that hasn't been pushed", func() {
			It("Produces no servingModelArtifacts", func() {
				incorrectArtifact := artifact()
				incorrectArtifact.Metadata.Fields["pushed"] = structpb.NewNumberValue(0)

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"an-artifact": {
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
									"an-artifact": {
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
				secondArtifact.Uri = "gs://some/where/else"

				Expect(modelServingArtifactsForJob(&aiplatformpb.PipelineJob{
					JobDetail: &aiplatformpb.PipelineJobDetail{
						TaskDetails: []*aiplatformpb.PipelineTaskDetail{
							{
								TaskName: "Pusher",
								Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
									"pushed_model": {
										Artifacts: []*aiplatformpb.Artifact{
											firstArtifact,
											secondArtifact,
										},
									},
								},
							},
						},
					},
				})).To(ConsistOf(common.Artifact{Name: "pushed_model", Location: "gs://some/where"}, common.Artifact{Name: "pushed_model", Location: "gs://some/where/else"}))
			})
		})
	})

	DescribeTable("toRunCompletionEvent for job that has completed", func(pipelineState aiplatformpb.PipelineState, status common.RunCompletionStatus) {
		runConfigurationName := common.RandomNamespacedName()
		pipelineName := common.RandomNamespacedName()
		pipelineRunName := common.RandomNamespacedName()
		artifactDefs := []pipelinesv1.Artifact{{
			Name: "an_artifact",
			Path: pipelinesv1.ArtifactPathDefinition{
				Path: pipelinesv1.ArtifactPath{
					Component: "a_component",
					Artifact:  "an_artifact",
				},
				Filter: "some == thing",
			},
		}}

		Expect(toRunCompletionEvent(&aiplatformpb.PipelineJob{
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
						TaskName: "Pusher",
						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
							"pushed_model": {
								Artifacts: []*aiplatformpb.Artifact{
									artifact(),
								},
							},
						},
					},
					{
						TaskName: "a_component",
						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
							"an_artifact": {
								Artifacts: []*aiplatformpb.Artifact{
									{
										DisplayName: "an_artifact",
										Uri:         "gs://some/where/else",
										Metadata: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"some": structpb.NewStringValue("thing"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, VAIRun{RunId: pipelineRunName.Name, Artifacts: artifactDefs})).To(Equal(&common.RunCompletionEvent{
			RunConfigurationName: runConfigurationName.NonEmptyPtr(),
			PipelineName:         pipelineName,
			RunName:              pipelineRunName.NonEmptyPtr(),
			RunId:                pipelineRunName.Name,
			Status:               status,
			ServingModelArtifacts: []common.Artifact{
				{
					Name:     "pushed_model",
					Location: "gs://some/where",
				},
			},
			Artifacts: []common.Artifact{
				{
					Name:     "an_artifact",
					Location: "gs://some/where/else",
				},
			},
		}))
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED, common.RunCompletionStatuses.Succeeded),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, common.RunCompletionStatuses.Failed),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED, common.RunCompletionStatuses.Failed),
	)

	Describe("runCompletionEventForRun", func() {
		When("GetPipelineJob errors", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("an error"))
				event := eventingServer.runCompletionEventForRun(context.Background(), VAIRun{
					RunId: common.RandomString(),
				})
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob return no result", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), VAIRun{
					RunId: common.RandomString(),
				})
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(&aiplatformpb.PipelineJob{
					State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
				}, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), VAIRun{
					RunId: common.RandomString(),
				})
				Expect(event).NotTo(BeNil())
			})
		})
	})
})
