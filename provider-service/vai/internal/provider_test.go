//go:build unit

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/golang/mock/gomock"
	"github.com/googleapis/gax-go/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockJobBuilder struct{ mock.Mock }

func (m *MockJobBuilder) MkRunPipelineJob(rd resource.RunDefinition) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rd)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

func (m *MockJobBuilder) MkRunSchedulePipelineJob(rsd resource.RunScheduleDefinition) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rsd)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

func (m *MockJobBuilder) MkSchedule(rsd resource.RunScheduleDefinition, pipelineJob *aiplatformpb.PipelineJob, parent string, maxConcurrentRunCount int64) (*aiplatformpb.Schedule, error) {
	args := m.Called(rsd, pipelineJob, parent, maxConcurrentRunCount)
	var schedule *aiplatformpb.Schedule
	if arg0 := args.Get(0); arg0 != nil {
		schedule = arg0.(*aiplatformpb.Schedule)
	}
	return schedule, args.Error(1)
}

type MockJobEnricher struct{ mock.Mock }

func (m *MockJobEnricher) Enrich(job *aiplatformpb.PipelineJob, raw map[string]any) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(job, raw)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg0 := args.Get(0); arg0 != nil {
		pipelineJob = arg0.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

type MockPipelineJobClient2 struct {
	mock.Mock
}

func(m *MockPipelineJobClient2) CreatePipelineJob(
	ctx context.Context,
	req *aiplatformpb.CreatePipelineJobRequest,
	opts ...gax.CallOption,
) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(ctx, req, opts)
	var pipelineJob *aiplatformpb.PipelineJob
	if arg1 := args.Get(0); arg1 != nil {
		pipelineJob = arg1.(*aiplatformpb.PipelineJob)
	}
	return pipelineJob, args.Error(1)
}

type MockFileHandler struct{ mock.Mock }

func (m *MockFileHandler) Write(p []byte, bucket string, filePath string) error {
	args := m.Called(p, bucket, filePath)
	return args.Error(0)
}

func (m *MockFileHandler) Delete(id string, bucket string) error {
	args := m.Called(id, bucket)
	return args.Error(0)
}

func (m *MockFileHandler) Read(bucket string, filePath string) (map[string]any, error) {
	args := m.Called(bucket, filePath)
	var data map[string]any
	if arg0 := args.Get(0); arg0 != nil {
		data = arg0.(map[string]any)
	}
	return data, args.Error(1)
}

// TODO extract to somewhere common
func randomPipelineDefinition() resource.PipelineDefinition {
	return resource.PipelineDefinition{
		Name:          common.RandomNamespacedName(),
		Version:       common.RandomString(),
		Image:         common.RandomString(),
		TfxComponents: common.RandomString(),
		Env:           make([]apis.NamedValue, 0),
		BeamArgs:      make([]apis.NamedValue, 0),
		Manifest:      json.RawMessage{},
	}
}

var _ = Describe("Provider", func() {
	var (
		mockFileHandler    MockFileHandler
		mockPipelineClient MockPipelineJobClient2
		mockScheduleClient *MockScheduleClient
		mockJobBuilder     MockJobBuilder
		mockJobEnricher    MockJobEnricher
		vaiProvider        VAIProvider
	)

	BeforeEach(func() {
		mockFileHandler = MockFileHandler{}
		mockPipelineClient = MockPipelineJobClient2{}
		mockScheduleClient = NewMockScheduleClient(mockCtrl)
		mockJobBuilder = MockJobBuilder{}
		mockJobEnricher = MockJobEnricher{}
		vaiProvider = VAIProvider{
			ctx:            context.Background(),
			config:         VAIProviderConfig{},
			fileHandler:    &mockFileHandler,
			pipelineClient: &mockPipelineClient,
			scheduleClient: mockScheduleClient,
			jobBuilder:     &mockJobBuilder,
			jobEnricher:    &mockJobEnricher,
		}
	})

	Context("CreatePipeline", func() {
		When("creating a pipeline", func() {
			It("should return the pipeline ID", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				pipelineDefinition := randomPipelineDefinition()
				pipelineId, err := vaiProvider.CreatePipeline(pipelineDefinition)

				Expect(err).ToNot(HaveOccurred())
				Expect(pipelineId).To(Equal(fmt.Sprintf("%s/%s", pipelineDefinition.Name.Namespace, pipelineDefinition.Name.Name)))
			})

			It("return an error when the file handler write fails", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed"))

				pipelineDefinition := randomPipelineDefinition()
				_, err := vaiProvider.CreatePipeline(pipelineDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("UpdatePipeline", func() {
		When("updating a pipeline", func() {
			It("should return the pipeline ID", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				pipelineDefinition := randomPipelineDefinition()
				pipelineId, err := vaiProvider.UpdatePipeline(pipelineDefinition, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(pipelineId).To(Equal(fmt.Sprintf("%s/%s", pipelineDefinition.Name.Namespace, pipelineDefinition.Name.Name)))
			})

			It("return an error when the file handler write fails", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed"))

				pipelineDefinition := randomPipelineDefinition()
				_, err := vaiProvider.CreatePipeline(pipelineDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("DeletePipeline", func() {
		When("deleting a pipeline", func() {
			It("should not return an error", func() {
				mockFileHandler.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				err := vaiProvider.DeletePipeline("")
				Expect(err).ToNot(HaveOccurred())
			})

			It("return an error when the file handler delete fails", func() {
				mockFileHandler.On("Delete", "pipelineId", mock.Anything, mock.Anything).Return(errors.New("failed"))
				err := vaiProvider.DeletePipeline("pipelineId")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("CreateRun", func() {
		When("creating a run", func() {
			It("return a run ID", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, nil)
				mockJobEnricher.On("Enrich", &pipelineJob, map[string]any{}).Return(&pipelineJob, nil)
				mockPipelineClient.On("CreatePipelineJob", mock.Anything, mock.Anything, mock.Anything).Return(&pipelineJob, nil)
				runId, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).ToNot(HaveOccurred())
				Expect(runId).To(Equal(fmt.Sprintf("%s-%s", runDefinition.Name.Namespace, runDefinition.Name.Name)))
			})

			It("return an error when the file handler read fails", func() {
				runDefinition := randomBasicRunDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, nil)
				mockJobEnricher.On("Enrich", &pipelineJob, map[string]any{}).Return(&pipelineJob, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the pipeline client fails", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, nil)
				mockJobEnricher.On("Enrich", &pipelineJob, map[string]any{}).Return(&pipelineJob, nil)
				mockPipelineClient.On("CreatePipelineJob", mock.Anything, mock.Anything, mock.Anything).Return(&pipelineJob, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})

	Context("CreateRunSchedule", func() {
		When("creating a run schedule", func() {
			It("return a schedule name", func() {
				runDefinition := randomRunScheduleDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunSchedulePipelineJob", runDefinition).Return(&pipelineJob, nil)
				mockJobEnricher.On("Enrich", &pipelineJob, map[string]any{}).Return(&pipelineJob, nil)
				mockPipelineClient.On("CreatePipelineJob", mock.Anything, mock.Anything, mock.Anything).Return(&pipelineJob, nil)
				mockJobBuilder.On("MkRunSchedule", mock.Anything, &pipelineJob, mock.Anything, mock.Anything).Return(&aiplatformpb.Schedule{}, nil)
				runId, err := vaiProvider.CreateRunSchedule(runDefinition)

				Expect(err).ToNot(HaveOccurred())
				Expect(runId).To(Equal(fmt.Sprintf("%s-%s", runDefinition.Name.Namespace, runDefinition.Name.Name)))
			})

			It("return an error when the file handler read fails", func() {
				runDefinition := randomBasicRunDefinition()
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job builder fails", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the job enricher fails", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, nil)
				mockJobEnricher.On("Enrich", &pipelineJob, map[string]any{}).Return(&pipelineJob, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})

			It("return an error when the pipeline client fails", func() {
				runDefinition := randomBasicRunDefinition()
				pipelineJob := aiplatformpb.PipelineJob{}
				mockFileHandler.On("Read", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
				mockJobBuilder.On("MkRunPipelineJob", runDefinition).Return(&pipelineJob, nil)
				mockJobEnricher.On("Enrich", &pipelineJob, map[string]any{}).Return(&pipelineJob, nil)
				mockPipelineClient.On("CreatePipelineJob", mock.Anything, mock.Anything, mock.Anything).Return(&pipelineJob, errors.New("failed"))
				_, err := vaiProvider.CreateRun(runDefinition)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})
})
