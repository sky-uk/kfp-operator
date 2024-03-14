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
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func artifact() *aiplatformpb.Artifact {
	return &aiplatformpb.Artifact{
		SchemaTitle: "tfx.PushedModel", // Legacy
		DisplayName: "a-model",
		Uri:         "gs://some/where",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"x": structpb.NewStructValue(
					&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"y": structpb.NewNumberValue(1),
						},
					}),
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

	DescribeTable("toRunCompletionEvent for job that has not completed", func(state aiplatformpb.PipelineState) {
		Expect(eventingServer.toRunCompletionEvent(&aiplatformpb.PipelineJob{State: state}, common.RandomString(), nil)).To(BeNil())
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_UNSPECIFIED),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_QUEUED),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_PENDING),
		Entry("Running", aiplatformpb.PipelineState_PIPELINE_STATE_RUNNING),
		Entry("Cancelling", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLING),
		Entry("Paused", aiplatformpb.PipelineState_PIPELINE_STATE_PAUSED),
	)

	DescribeTable("toRunCompletionEvent for job that has completed", func(pipelineState aiplatformpb.PipelineState, status common.RunCompletionStatus) {
		runConfigurationName := common.RandomNamespacedName()
		pipelineName := common.RandomNamespacedName()
		pipelineRunName := common.RandomNamespacedName()

		Expect(eventingServer.toRunCompletionEvent(&aiplatformpb.PipelineJob{
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
		}, pipelineRunName.Name, nil)).To(Equal(&common.RunCompletionEvent{
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
		}))
	},
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED, common.RunCompletionStatuses.Succeeded),
		Entry("Unspecified", aiplatformpb.PipelineState_PIPELINE_STATE_FAILED, common.RunCompletionStatuses.Failed),
		Entry("Pending", aiplatformpb.PipelineState_PIPELINE_STATE_CANCELLED, common.RunCompletionStatuses.Failed),
	)

	Describe("artifactsForJob", func() {
		When("The job is missing the component", func() {
			It("Produces no artifacts", func() {
				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
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

		When("The job is missing artifact index", func() {
			It("Produces no artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
							Component: componentName,
							Artifact:  outputName,
							Index:     1,
						},
					},
				}}

				incorrectArtifact := artifact()

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

		When("The job has a matching component that misses the output", func() {
			It("Produces no artifacts", func() {
				componentName := common.RandomString()
				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
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

		When("The job has a matching path but the artifact has no uri", func() {
			It("Produces no artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				incorrectArtifact := artifact()
				incorrectArtifact.Uri = ""

				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
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

		When("The job has a matching path but no matching properties", func() {
			It("Produces no artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				incorrectArtifact := artifact()
				incorrectArtifact.Metadata.Fields = nil

				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: common.RandomString(),
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
							Component: componentName,
							Artifact:  outputName,
							Index:     0,
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

		When("The job has matching artifacts", func() {
			It("Produces the artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				artifactName := common.RandomString()
				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: artifactName,
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
							Component: componentName,
							Artifact:  outputName,
							Index:     1,
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
				}, artifactDefs)).To(ConsistOf(common.Artifact{Name: artifactName, Location: "gs://some/where/else"}))
			})
		})

		When("The job has matching artifacts with matching properties", func() {
			It("Produces the artifacts", func() {
				componentName := common.RandomString()
				outputName := common.RandomString()
				artifactName := common.RandomString()
				artifactDefs := []pipelinesv1.OutputArtifact{{
					Name: artifactName,
					Path: pipelinesv1.ArtifactPath{
						Locator: pipelinesv1.ArtifactLocator{
							Component: componentName,
							Artifact:  outputName,
							Index:     1,
						},
						Filter: "x.y == 1",
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
				}, artifactDefs)).To(ConsistOf(common.Artifact{Name: artifactName, Location: "gs://some/where/else"}))
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

	Describe("runCompletionEventForRun", func() {
		When("GetPipelineJob errors", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("an error"))
				event := eventingServer.runCompletionEventForRun(context.Background(), common.RandomString(), nil)
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob return no result", func() {
			It("returns no event", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(nil, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), common.RandomString(), nil)
				Expect(event).To(BeNil())
			})
		})

		When("GetPipelineJob returns a PipelineJob", func() {
			It("Returns a RunCompletionEvent", func() {
				mockPipelineJobClient.EXPECT().GetPipelineJob(gomock.Any(), gomock.Any()).Return(&aiplatformpb.PipelineJob{
					State: aiplatformpb.PipelineState_PIPELINE_STATE_SUCCEEDED,
				}, nil)
				event := eventingServer.runCompletionEventForRun(context.Background(), common.RandomString(), nil)
				Expect(event).NotTo(BeNil())
			})
		})
	})

	DescribeTable("gvrForRunLabels", func(labels map[string]string, expectedGvr *schema.GroupVersionResource, expectedNamespacedName *types.NamespacedName) {
		gvr, namespacedName, err := gvrAndNamespacedNameForRunLabels(labels)

		if expectedGvr == nil {
			Expect(err).To(HaveOccurred())
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(gvr).To(Equal(*expectedGvr))
			Expect(namespacedName).To(Equal(*expectedNamespacedName))
		}

	},
		Entry(nil, map[string]string{labels.RunConfigurationName: "rc"}, &base.RunConfigurationGVR, &types.NamespacedName{Name: "rc"}),
		Entry(nil, map[string]string{labels.RunConfigurationName: "rc", labels.RunConfigurationNamespace: "rcNamespace"}, &base.RunConfigurationGVR, &types.NamespacedName{Name: "rc", Namespace: "rcNamespace"}),
		Entry(nil, map[string]string{labels.RunConfigurationName: "run"}, &base.RunConfigurationGVR, &types.NamespacedName{Name: "run"}),
		Entry(nil, map[string]string{labels.RunConfigurationName: "run", labels.RunConfigurationNamespace: "runNamespace"}, &base.RunConfigurationGVR, &types.NamespacedName{Name: "run", Namespace: "runNamespace"}),
		Entry(nil, map[string]string{}, nil, nil))
})
