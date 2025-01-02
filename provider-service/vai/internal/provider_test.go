//go:build unit

package internal

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

type MockJobBuilder struct{ mock.Mock }

func (m MockJobBuilder) MkRunPipelineJob(rd resource.RunDefinition) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rd)
	pipelineJob := args.Get(0).(aiplatformpb.PipelineJob)
	return &pipelineJob, args.Error(1)
}

func (m MockJobBuilder) MkRunSchedulePipelineJob(rsd resource.RunScheduleDefinition) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(rsd)
	pipelineJob := args.Get(0).(aiplatformpb.PipelineJob)
	return &pipelineJob, args.Error(1)
}

func (m MockJobBuilder) MkSchedule(rsd resource.RunScheduleDefinition, pipelineJob *aiplatformpb.PipelineJob, parent string, maxConcurrentRunCount int64) (*aiplatformpb.Schedule, error) {
	args := m.Called(rsd, pipelineJob, parent, maxConcurrentRunCount)
	schedule := args.Get(0).(aiplatformpb.Schedule)
	return &schedule, args.Error(1)
}

type MockJobEnricher struct{ mock.Mock }

func (m MockJobEnricher) Enrich(job *aiplatformpb.PipelineJob, raw map[string]any) (*aiplatformpb.PipelineJob, error) {
	args := m.Called(job, raw)
	pipelineJob := args.Get(0).(aiplatformpb.PipelineJob)
	return &pipelineJob, args.Error(1)
}

type MockFileHandler struct{ mock.Mock }

func (m MockFileHandler) Write(p []byte, bucket string, filePath string) error {
	args := m.Called(p, bucket, filePath)
	return args.Error(0)
}

func (m MockFileHandler) Delete(id string, bucket string) error {
	args := m.Called(id, bucket)
	return args.Error(0)
}

func (m MockFileHandler) Read(bucket string, filePath string) (map[string]any, error) {
	args := m.Called(bucket, filePath)
	return args.Get(0).(map[string]any), args.Error(1)
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
		mockCtrl           *gomock.Controller
		mockFileHandler    MockFileHandler
		mockPipelineClient *MockPipelineJobClient
		mockScheduleClient *MockScheduleClient
		mockJobBuilder     MockJobBuilder
		mockJobEnricher    MockJobEnricher
		vaiProvider        VAIProvider
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockFileHandler = MockFileHandler{}
		mockPipelineClient = NewMockPipelineJobClient(mockCtrl)
		mockScheduleClient = NewMockScheduleClient(mockCtrl)
		vaiProvider = VAIProvider{
			ctx:            context.Background(),
			config:         VAIProviderConfig{},
			fileHandler:    &mockFileHandler,
			pipelineClient: mockPipelineClient,
			scheduleClient: mockScheduleClient,
			jobBuilder:     &mockJobBuilder,
			jobEnricher:    &mockJobEnricher,
		}
	})

	Context("CreatePipeline", func() {
		When("creating a pipeline", func() {
			It("should return the pipeline ID when it succeeds", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				pipelineDefinition := randomPipelineDefinition()
				pipelineId, err := vaiProvider.CreatePipeline(pipelineDefinition)

				Expect(err).ToNot(HaveOccurred())
				Expect(pipelineId).To(Equal(fmt.Sprintf("%s/%s", pipelineDefinition.Name.Namespace, pipelineDefinition.Name.Name)))
			})

			It("return an error when it fails", func() {
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
			It("should return the pipeline ID when it succeeds", func() {
				mockFileHandler.On("Write", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				pipelineDefinition := randomPipelineDefinition()
				pipelineId, err := vaiProvider.UpdatePipeline(pipelineDefinition, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(pipelineId).To(Equal(fmt.Sprintf("%s/%s", pipelineDefinition.Name.Namespace, pipelineDefinition.Name.Name)))
			})

			It("return an error when it fails", func() {
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
			It("should not return an error when it succeeds", func() {
				mockFileHandler.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				err := vaiProvider.DeletePipeline("")
				Expect(err).ToNot(HaveOccurred())
			})

			It("return an error when it fails", func() {
				mockFileHandler.On("Delete", "pipelineId", mock.Anything, mock.Anything).Return(errors.New("failed"))
				err := vaiProvider.DeletePipeline("pipelineId")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("failed"))
			})
		})
	})
})
