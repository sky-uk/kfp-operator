//go:build unit

package resource

import (
	"encoding/json"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Pipeline", Ordered, func() {
	var (
		mockProvider MockPipelineProvider
		underTest    Pipeline
	)

	BeforeEach(func() {
		mockProvider = MockPipelineProvider{}
		underTest = Pipeline{Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				pipeline := PipelineDefinitionWrapper{}

				jsonPipeline, err := json.Marshal(pipeline)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"

				mockProvider.On("CreatePipeline", pipeline).Return(id, nil)

				response, err := underTest.Create(jsonPipeline)

				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)

				response, err := underTest.Create(invalidJson)

				Expect(err).To(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				pipeline := PipelineDefinitionWrapper{}
				jsonPipeline, err := json.Marshal(pipeline)
				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")

				mockProvider.On("CreatePipeline", pipeline).Return("", expectedErr)

				response, err := underTest.Create(jsonPipeline)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(ResponseBody{Id: ""}))
			})
		})
	})

	Context("Update", func() {
		When("valid json passed, and provider operations succeed", func() {
			It("returns no error", func() {
				pipeline := PipelineDefinitionWrapper{}

				jsonPipeline, err := json.Marshal(pipeline)
				Expect(err).ToNot(HaveOccurred())

				id := "some-id"

				mockProvider.On("UpdatePipeline", pipeline, id).Return(id, nil)

				err = underTest.Update(id, jsonPipeline)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				err := underTest.Update("some-id", invalidJson)
				Expect(err).To(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				pipeline := PipelineDefinitionWrapper{}
				jsonExperiment, err := json.Marshal(pipeline)
				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")

				id := "some-id"

				mockProvider.On("UpdatePipeline", pipeline, id).Return("", expectedErr)

				err = underTest.Update(id, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeletePipeline", id).Return(nil)

				err := underTest.Delete(id)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeletePipeline", id).Return(expectedErr)

				err := underTest.Delete(id)
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockPipelineProvider struct {
	mock.Mock
}

func (m *MockPipelineProvider) CreatePipeline(pd PipelineDefinitionWrapper) (string, error) {
	args := m.Called(pd)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineProvider) UpdatePipeline(pd PipelineDefinitionWrapper, id string) (string, error) {
	args := m.Called(pd, id)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineProvider) DeletePipeline(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
