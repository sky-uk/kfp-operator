//go:build unit

package resource

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Pipeline", Ordered, func() {
	var (
		mockProvider MockPipelineProvider
		p            Pipeline
		ctx          = context.Background()
	)

	ignoreCtx := mock.Anything

	BeforeEach(func() {
		mockProvider = MockPipelineProvider{}
		p = Pipeline{Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				pdw := PipelineDefinitionWrapper{}
				jsonPipeline, err := json.Marshal(pdw)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				mockProvider.On("CreatePipeline", ignoreCtx, pdw).Return(id, nil)
				resp, err := p.Create(ctx, jsonPipeline)

				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(base.Output{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				response, err := p.Create(ctx, invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(response).To(Equal(base.Output{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				pdw := PipelineDefinitionWrapper{}
				jsonPipeline, err := json.Marshal(pdw)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				mockProvider.On("CreatePipeline", ignoreCtx, pdw).Return("", expectedErr)
				response, err := p.Create(ctx, jsonPipeline)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(base.Output{}))
			})
		})
	})

	Context("Update", func() {
		When("valid json passed, and provider operations succeed", func() {
			It("returns no error", func() {
				pdw := PipelineDefinitionWrapper{}
				jsonPipeline, err := json.Marshal(pdw)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				updatedId := "some-update-id"
				mockProvider.On("UpdatePipeline", ignoreCtx, pdw, id).Return(updatedId, nil)
				resp, err := p.Update(ctx, id, jsonPipeline)

				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal(base.Output{Id: updatedId}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				resp, err := p.Update(ctx, "some-id", invalidJson)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(resp).To(Equal(base.Output{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				pdw := PipelineDefinitionWrapper{}
				jsonExperiment, err := json.Marshal(pdw)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				id := "some-id"
				mockProvider.On("UpdatePipeline", ignoreCtx, pdw, id).Return("", expectedErr)
				resp, err := p.Update(ctx, id, jsonExperiment)

				Expect(err).To(Equal(expectedErr))
				Expect(resp).To(Equal(base.Output{}))
			})
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeletePipeline", ignoreCtx, id).Return(nil)
				err := p.Delete(ctx, id)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeletePipeline", ignoreCtx, id).Return(expectedErr)
				err := p.Delete(ctx, id)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockPipelineProvider struct {
	mock.Mock
}

func (m *MockPipelineProvider) CreatePipeline(
	ctx context.Context,
	pdw PipelineDefinitionWrapper,
) (string, error) {
	args := m.Called(ctx, pdw)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineProvider) UpdatePipeline(
	ctx context.Context,
	pdw PipelineDefinitionWrapper,
	id string,
) (string, error) {
	args := m.Called(ctx, pdw, id)
	return args.String(0), args.Error(1)
}

func (m *MockPipelineProvider) DeletePipeline(
	ctx context.Context,
	id string,
) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
