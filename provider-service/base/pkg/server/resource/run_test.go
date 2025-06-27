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

var _ = Describe("Run", Ordered, func() {
	var (
		mockProvider MockRunProvider
		r            Run
		ctx          = context.Background()
	)

	ignoreCtx := mock.Anything

	BeforeEach(func() {
		mockProvider = MockRunProvider{}
		r = Run{Provider: &mockProvider}
	})

	Context("Create", func() {
		When("valid json passed, and provider returns success", func() {
			It("returns the id of the resource", func() {
				rd := RunDefinition{}
				jsonRun, err := json.Marshal(rd)

				Expect(err).ToNot(HaveOccurred())

				id := "some-id"
				mockProvider.On("CreateRun", ignoreCtx, rd).Return(id, nil)
				response, err := r.Create(ctx, jsonRun, nil)

				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(Equal(ResponseBody{Id: id}))
			})
		})

		When("invalid json is passed", func() {
			It("errors", func() {
				invalidJson := []byte(`/n`)
				response, err := r.Create(ctx, invalidJson, nil)

				var expectedErr *UserError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
				Expect(response).To(Equal(ResponseBody{}))
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				rd := RunDefinition{}
				jsonRun, err := json.Marshal(rd)

				Expect(err).ToNot(HaveOccurred())

				expectedErr := errors.New("some-error")
				mockProvider.On("CreateRun", ignoreCtx, rd).Return("", expectedErr)
				response, err := r.Create(ctx, jsonRun, nil)

				Expect(err).To(Equal(expectedErr))
				Expect(response).To(Equal(ResponseBody{}))
			})
		})
	})

	Context("Update", func() {
		It("returns an Unimplemented error", func() {
			id := "some-id"
			response, err := r.Update(ctx, id, []byte("foo"), nil)

			var expectedErr *UnimplementedError
			Expect(errors.As(err, &expectedErr)).To(BeTrue())
			Expect(response).To(Equal(ResponseBody{}))
		})
	})

	Context("Delete", func() {
		When("valid id is passed and provider operations succeed", func() {
			It("return no error", func() {
				id := "some-id"
				mockProvider.On("DeleteRun", ignoreCtx, id).Return(nil)
				err := r.Delete(ctx, id, nil)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("provider errors", func() {
			It("errors", func() {
				id := "some-id"
				expectedErr := errors.New("some-error")
				mockProvider.On("DeleteRun", ignoreCtx, id).Return(expectedErr)
				err := r.Delete(ctx, id, nil)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})

type MockRunProvider struct {
	mock.Mock
}

func (m *MockRunProvider) CreateRun(ctx context.Context, rd RunDefinition) (string, error) {
	args := m.Called(ctx, rd)
	return args.String(0), args.Error(1)
}

func (m *MockRunProvider) DeleteRun(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
