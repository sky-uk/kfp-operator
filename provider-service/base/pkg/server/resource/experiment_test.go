//go:build unit

package resource

import (
	"encoding/json"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Experiment", Ordered, func() {
	var (
		mockProvider MockExperimentProvider
		underTest    Experiment
	)

	BeforeEach(func() {
		mockProvider = MockExperimentProvider{}
		underTest = Experiment{Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				experiment := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(experiment)
				Expect(err).ToNot(HaveOccurred())

				id := "some-id"

				mockProvider.On("CreateExperiment", experiment).Return(id, nil)
				response, err := underTest.Create(jsonExperiment)
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
				experiment := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(experiment)
				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")

				mockProvider.On("CreateExperiment", experiment).Return("", expectedErr)

				response, err := underTest.Create(jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(ResponseBody{Id: ""}))
			})
		})
	})

	Context("Update", func() {
		When("valid json passed, and provider operations succeed", func() {
			It("returns no error", func() {
				experiment := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(experiment)
				Expect(err).ToNot(HaveOccurred())

				id := "some-id"

				mockProvider.On("UpdateExperiment", experiment, id).Return(id, nil)

				err = underTest.Update(id, jsonExperiment)
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
				experiment := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(experiment)
				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")

				id := "some-id"

				mockProvider.On("UpdateExperiment", experiment, id).Return("", expectedErr)

				err = underTest.Update(id, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeleteExperiment", id).Return(nil)

				err := underTest.Delete(id)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeleteExperiment", id).Return(expectedErr)

				err := underTest.Delete(id)
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockExperimentProvider struct {
	mock.Mock
}

func (m *MockExperimentProvider) CreateExperiment(ed ExperimentDefinition) (string, error) {
	args := m.Called(ed)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentProvider) UpdateExperiment(ed ExperimentDefinition, id string) (string, error) {
	args := m.Called(ed, id)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentProvider) DeleteExperiment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
