//go:build unit

package vai

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"google.golang.org/protobuf/types/known/structpb"
)

func artifact() *aiplatformpb.Artifact {
	return &aiplatformpb.Artifact{
		SchemaTitle: "tfx.PushedModel", // Legacy
		DisplayName: "a-model",
		Uri:         "gs://some/where",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				//"x": structpb.NewStructValue(
				//	&structpb.Struct{
				//		Fields: map[string]*structpb.Value{
				//			"y": structpb.NewNumberValue(1),
				//		},
				//	}),
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
		eventingServer        VaiEventingServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockPipelineJobClient = NewMockPipelineJobClient(mockCtrl)
		eventingServer = VaiEventingServer{
			ProviderConfig: VAIProviderConfig{
				Name: common.RandomString(),
			},
			PipelineJobClient: mockPipelineJobClient,
			Logger:            logr.Discard(),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	DescribeTable("toRunCompletionEventData for job that has not completed", func(state aiplatformpb.PipelineState) {
		Expect(eventingServer.toRunCompletionEventData(&aiplatformpb.PipelineJob{State: state}, common.RandomString())).To(BeNil())
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

		Expect(eventingServer.toRunCompletionEventData(&aiplatformpb.PipelineJob{
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
			Provider: eventingServer.ProviderConfig.Name,
			ComponentCompletion: []common.PipelineComponent{
				{
					Name: "my-task-name",
					ComponentArtifactDetails: []common.ComponentArtifactDetails{
						{
							ArtifactName: "a-model",
							Artifacts: []common.ComponentOutputArtifact{
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

		// move to KFP Operator
		//When("The job is missing artifact index", func() {
		//	It("Produces no artifacts", func() {
		//		componentName := common.RandomString()
		//		outputName := common.RandomString()
		//
		//		incorrectArtifact := artifact()
		//
		//		Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
		//			JobDetail: &aiplatformpb.PipelineJobDetail{
		//				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
		//					{
		//						TaskName: componentName,
		//						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
		//							outputName: {
		//								Artifacts: []*aiplatformpb.Artifact{
		//									incorrectArtifact,
		//								},
		//							},
		//						},
		//					},
		//				},
		//			},
		//		})).To(BeEmpty())
		//	})
		//})

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
						Name:                     componentName,
						ComponentArtifactDetails: []common.ComponentArtifactDetails{},
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
						ComponentArtifactDetails: []common.ComponentArtifactDetails{
							{
								ArtifactName: "a-model",
								Artifacts:    []common.ComponentOutputArtifact{},
							},
						},
					},
				}))
			})
		})

		// to move to KFP Operator
		//When("The job has a matching path but the artifact has no uri", func() {
		//	It("Produces no artifacts", func() {
		//		componentName := common.RandomString()
		//		outputName := common.RandomString()
		//		incorrectArtifact := artifact()
		//		incorrectArtifact.Uri = ""
		//
		//		Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
		//			JobDetail: &aiplatformpb.PipelineJobDetail{
		//				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
		//					{
		//						TaskName: componentName,
		//						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
		//							outputName: {
		//								Artifacts: []*aiplatformpb.Artifact{
		//									incorrectArtifact,
		//								},
		//							},
		//						},
		//					},
		//				},
		//			},
		//		})).To(BeEmpty())
		//	})
		//})

		//When("The job has a matching path but no matching properties", func() {
		//	It("Produces no artifacts", func() {
		//		componentName := common.RandomString()
		//		outputName := common.RandomString()
		//		incorrectArtifact := artifact()
		//		incorrectArtifact.Metadata.Fields = nil
		//
		//		Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
		//			JobDetail: &aiplatformpb.PipelineJobDetail{
		//				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
		//					{
		//						TaskName: componentName,
		//						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
		//							outputName: {
		//								Artifacts: []*aiplatformpb.Artifact{
		//									artifact(),
		//								},
		//							},
		//						},
		//					},
		//				},
		//			},
		//		})).To(BeEmpty())
		//	})
		//})

		//When("The job has matching artifacts", func() {
		//	It("Produces the artifacts", func() {
		//		componentName := common.RandomString()
		//		outputName := common.RandomString()
		//		artifactName := common.RandomString()
		//
		//		firstArtifact := artifact()
		//		secondArtifact := artifact()
		//		secondArtifact.Uri = "gs://some/where/else"
		//		secondArtifact.DisplayName = "another-artifact"
		//
		//		Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
		//			JobDetail: &aiplatformpb.PipelineJobDetail{
		//				TaskDetails: []*aiplatformpb.PipelineTaskDetail{
		//					{
		//						TaskName: componentName,
		//						Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
		//							outputName: {
		//								Artifacts: []*aiplatformpb.Artifact{
		//									firstArtifact,
		//									secondArtifact,
		//								},
		//							},
		//						},
		//					},
		//				},
		//			},
		//		})).To(ConsistOf(common.Artifact{Name: artifactName, Location: "gs://some/where/else"}))
		//	})
		//})

		//	When("The job has matching artifacts with matching properties", func() {
		//		It("Produces the artifacts", func() {
		//			componentName := common.RandomString()
		//			outputName := common.RandomString()
		//			artifactName := common.RandomString()
		//
		//			firstArtifact := artifact()
		//			secondArtifact := artifact()
		//			secondArtifact.Uri = "gs://some/where/else"
		//			secondArtifact.DisplayName = "another-artifact"
		//
		//			Expect(artifactsFilterData(&aiplatformpb.PipelineJob{
		//				JobDetail: &aiplatformpb.PipelineJobDetail{
		//					TaskDetails: []*aiplatformpb.PipelineTaskDetail{
		//						{
		//							TaskName: componentName,
		//							Outputs: map[string]*aiplatformpb.PipelineTaskDetail_ArtifactList{
		//								outputName: {
		//									Artifacts: []*aiplatformpb.Artifact{
		//										firstArtifact,
		//										secondArtifact,
		//									},
		//								},
		//							},
		//						},
		//					},
		//				},
		//			})).To(ConsistOf(common.Artifact{Name: artifactName, Location: "gs://some/where/else"}))
		//		})
		//	})
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
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("an error"))
				event := eventingServer.runCompletionEventDataForRun(context.Background(), common.RandomString())
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob return no result", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, nil)
				event := eventingServer.runCompletionEventDataForRun(context.Background(), common.RandomString())
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(&aiplatformpb.PipelineJob{
					State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
				}, nil)
				event := eventingServer.runCompletionEventDataForRun(context.Background(), common.RandomString())
				Expect(event).NotTo(BeNil())
			})
		})
	})
})
