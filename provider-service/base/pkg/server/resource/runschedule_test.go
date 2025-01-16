//go:build unit

package resource

import (
	"encoding/json"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("RunSchedule", Ordered, func() {
	var (
		mockProvider MockRunScheduleProvider
		rs           RunSchedule
	)

	BeforeEach(func() {
		mockProvider = MockRunScheduleProvider{}
		rs = RunSchedule{Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				rsd := RunScheduleDefinition{}

				jsonRunSchedule, err := json.Marshal(rsd)
				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				mockProvider.On("CreateRunSchedule", rsd).Return(id, nil)

				response, err := rs.Create(jsonRunSchedule)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)

				response, err := rs.Create(invalidJson)
				Expect(err).To(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				rsd := RunScheduleDefinition{}

				jsonRunSchedule, err := json.Marshal(rsd)
				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				mockProvider.On("CreateRunSchedule", rsd).Return("", expectedErr)

				response, err := rs.Create(jsonRunSchedule)
				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(ResponseBody{Id: ""}))
			})
		})
	})

	Context("Update", func() {
		When("valid json passed, and provider operations succeed", func() {
			It("returns no error", func() {
				rsd := RunScheduleDefinition{}

				jsonRunSchedule, err := json.Marshal(rsd)
				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				updatedId := "some-update-id"
				mockProvider.On("UpdateRunSchedule", rsd, id).Return(updatedId, nil)

				resp, err := rs.Update(id, jsonRunSchedule)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Id).To(Equal(updatedId))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				resp, err := rs.Update("some-id", invalidJson)
				Expect(err).To(HaveOccurred())
				Expect(resp.Id).To(BeEmpty())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				rsd := RunScheduleDefinition{}
				jsonExperiment, err := json.Marshal(rsd)
				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				id := "some-id"
				mockProvider.On("UpdateRunSchedule", rsd, id).Return("", expectedErr)

				resp, err := rs.Update(id, jsonExperiment)
				Expect(err).To(Equal(expectedErr))
				Expect(resp.Id).To(BeEmpty())
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeleteRunSchedule", id).Return(nil)

				err := rs.Delete(id)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeleteRunSchedule", id).Return(expectedErr)

				err := rs.Delete(id)
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockRunScheduleProvider struct {
	mock.Mock
}

func (m *MockRunScheduleProvider) CreateRunSchedule(
	rsd RunScheduleDefinition,
) (string, error) {
	args := m.Called(rsd)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) UpdateRunSchedule(
	rsd RunScheduleDefinition,
	id string,
) (string, error) {
	args := m.Called(rsd, id)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) DeleteRunSchedule(rsd string) error {
	args := m.Called(rsd)
	return args.Error(0)
}
