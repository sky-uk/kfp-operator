//go:build unit

package resource

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Experiment", Ordered, func() {
	var (
		mockProvider MockExperimentProvider
		exp          Experiment
	)

	BeforeEach(func() {
		mockProvider = MockExperimentProvider{}
		exp = Experiment{Ctx: context.Background(), Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				ed := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				mockProvider.On("CreateExperiment", ed).Return(id, nil)
				response, err := exp.Create(jsonExperiment)

				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				response, err := exp.Create(invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(response).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				ed := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				mockProvider.On("CreateExperiment", ed).Return("", expectedErr)
				response, err := exp.Create(jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(ResponseBody{}))
			})
		})
	})

	Context("Update", func() {
		When("valid json passed, and provider operations succeed", func() {
			It("returns no error", func() {
				ed := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				updatedId := "some-update-id"
				mockProvider.On("UpdateExperiment", ed, id).Return(updatedId, nil)
				resp, err := exp.Update(id, jsonExperiment)

				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(ResponseBody{Id: updatedId}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				resp, err := exp.Update("some-id", invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(resp).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				ed := ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				id := "some-id"
				mockProvider.On("UpdateExperiment", ed, id).Return("", expectedErr)
				resp, err := exp.Update(id, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(resp).To(Equal(ResponseBody{}))
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeleteExperiment", id).Return(nil)
				err := exp.Delete(id)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeleteExperiment", id).Return(expectedErr)
				err := exp.Delete(id)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockExperimentProvider struct {
	mock.Mock
}

func (m *MockExperimentProvider) CreateExperiment(
	ed ExperimentDefinition,
) (string, error) {
	args := m.Called(ed)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentProvider) UpdateExperiment(
	ed ExperimentDefinition,
	id string,
) (string, error) {
	args := m.Called(ed, id)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentProvider) DeleteExperiment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
