//go:build unit

package resource

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Experiment", Ordered, func() {
	var (
		mockProvider MockExperimentProvider
		exp          Experiment
		ctx          = context.Background()
	)

	ignoreCtx := mock.Anything

	BeforeEach(func() {
		mockProvider = MockExperimentProvider{}
		exp = Experiment{Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				ed := base.ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				mockProvider.On("CreateExperiment", ignoreCtx, ed).Return(id, nil)
				response, err := exp.Create(ctx, jsonExperiment)

				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(base.Output{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				response, err := exp.Create(ctx, invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(response).To(Equal(base.Output{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				ed := base.ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				mockProvider.On("CreateExperiment", ignoreCtx, ed).Return("", expectedErr)
				response, err := exp.Create(ctx, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(base.Output{}))
			})
		})
	})

	Context("Update", func() {
		When("valid json passed, and provider operations succeed", func() {
			It("returns no error", func() {
				ed := base.ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				updatedId := "some-update-id"
				mockProvider.On("UpdateExperiment", ignoreCtx, ed, id).Return(updatedId, nil)
				resp, err := exp.Update(ctx, id, jsonExperiment)

				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(base.Output{Id: updatedId}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				resp, err := exp.Update(ctx, "some-id", invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(resp).To(Equal(base.Output{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				ed := base.ExperimentDefinition{}
				jsonExperiment, err := json.Marshal(ed)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				id := "some-id"
				mockProvider.On("UpdateExperiment", ignoreCtx, ed, id).Return("", expectedErr)
				resp, err := exp.Update(ctx, id, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(resp).To(Equal(base.Output{}))
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeleteExperiment", ignoreCtx, id).Return(nil)
				err := exp.Delete(ctx, id)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeleteExperiment", ignoreCtx, id).Return(expectedErr)
				err := exp.Delete(ctx, id)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockExperimentProvider struct {
	mock.Mock
}

func (m *MockExperimentProvider) CreateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
) (string, error) {
	args := m.Called(ctx, ed)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentProvider) UpdateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
	id string,
) (string, error) {
	args := m.Called(ctx, ed, id)
	return args.String(0), args.Error(1)
}

func (m *MockExperimentProvider) DeleteExperiment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
