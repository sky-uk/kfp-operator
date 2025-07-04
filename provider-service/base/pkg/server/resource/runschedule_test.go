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

var _ = Describe("RunSchedule", Ordered, func() {
	var (
		mockProvider MockRunScheduleProvider
		rs           RunSchedule
		ctx          = context.Background()
	)

	ignoreCtx := mock.Anything

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
				mockProvider.On("CreateRunSchedule", ignoreCtx, rsd).Return(id, nil)
				response, err := rs.Create(ctx, jsonRunSchedule)

				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				response, err := rs.Create(ctx, invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(response).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				rsd := RunScheduleDefinition{}
				jsonRunSchedule, err := json.Marshal(rsd)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				mockProvider.On("CreateRunSchedule", ignoreCtx, rsd).Return("", expectedErr)
				response, err := rs.Create(ctx, jsonRunSchedule)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(ResponseBody{}))
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
				mockProvider.On("UpdateRunSchedule", ignoreCtx, rsd, id).Return(updatedId, nil)
				resp, err := rs.Update(ctx, id, jsonRunSchedule)

				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(ResponseBody{Id: updatedId}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				resp, err := rs.Update(ctx, "some-id", invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(resp).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				rsd := RunScheduleDefinition{}
				jsonExperiment, err := json.Marshal(rsd)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				id := "some-id"
				mockProvider.On("UpdateRunSchedule", ignoreCtx, rsd, id).Return("", expectedErr)
				resp, err := rs.Update(ctx, id, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(resp).To(Equal(ResponseBody{}))
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeleteRunSchedule", ignoreCtx, id).Return(nil)
				err := rs.Delete(ctx, id)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeleteRunSchedule", ignoreCtx, id).Return(expectedErr)
				err := rs.Delete(ctx, id)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockRunScheduleProvider struct {
	mock.Mock
}

func (m *MockRunScheduleProvider) CreateRunSchedule(
	ctx context.Context,
	rsd RunScheduleDefinition,
) (string, error) {
	args := m.Called(ctx, rsd)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) UpdateRunSchedule(
	ctx context.Context,
	rsd RunScheduleDefinition,
	id string,
) (string, error) {
	args := m.Called(ctx, rsd, id)
	return args.String(0), args.Error(1)
}

func (m *MockRunScheduleProvider) DeleteRunSchedule(ctx context.Context, rsd string) error {
	args := m.Called(ctx, rsd)
	return args.Error(0)
}
