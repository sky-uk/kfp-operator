//go:build unit

package server

import (
	"bytes"
	"context"
	"errors"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/trigger"
	"testing"

	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/stretchr/testify/mock"
)

func TestServerUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Unit Test")
}

type MockHandledResource struct{ mock.Mock }

func (m *MockHandledResource) Type() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockHandledResource) Create(ctx context.Context, body []byte, headers map[string]string) (resource.ResponseBody, error) {
	args := m.Called(ctx, body, headers)
	var response resource.ResponseBody
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(resource.ResponseBody)
	}
	return response, args.Error(1)
}

func (m *MockHandledResource) Update(
	ctx context.Context,
	id string,
	body []byte,
	headers map[string]string,
) (resource.ResponseBody, error) {
	args := m.Called(ctx, id, body, headers)
	var response resource.ResponseBody
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(resource.ResponseBody)
	}
	return response, args.Error(1)
}

func (m *MockHandledResource) Delete(ctx context.Context, id string, headers map[string]string) error {
	args := m.Called(ctx, id, headers)
	return args.Error(0)
}

type failReader struct{}

func (f *failReader) Read(_ []byte) (int, error) {
	return 0, errors.New("simulated read error")
}

var _ = Describe("Http Server Endpoints", func() {
	var (
		server          *httptest.Server
		resourceType    = "mock-resource"
		payload         = []byte(`{"name": "test"}`)
		handledResource *MockHandledResource
		ctx             = context.Background()
	)

	ignoreCtx := mock.Anything

	BeforeEach(func() {
		handledResource = &MockHandledResource{}
		handledResource.On("Type").Return(resourceType)
		server = httptest.NewServer(newHandler(ctx, []resource.HttpHandledResource{
			handledResource,
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	Context("/readyz", func() {
		When("called", func() {
			It("should be OK", func() {
				resp, err := http.Get(server.URL + "/readyz")

				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)

				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal("Application is ready."))
			})
		})
	})

	Context("/livez", func() {
		When("called", func() {
			It("should be OK", func() {
				resp, err := http.Get(server.URL + "/livez")

				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)

				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal("Application is live."))
			})
		})
	})

	Context("/resource/{resource.Type()}", func() {
		Context("/ POST request createHandler", func() {
			When("succeeds", func() {
				It("returns 201 with valid response body", func() {
					response := "mocked-id"
					expectedHeaders := map[string]string{
						trigger.TriggerType:            "TestType",
						trigger.TriggerSource:          "TestSource",
						trigger.TriggerSourceNamespace: "TestNamespace",
					}

					handledResource.On("Create", ignoreCtx, payload, expectedHeaders).Return(
						resource.ResponseBody{
							Id: response,
						},
						nil,
					)

					req := httptest.NewRequest(
						http.MethodPost,
						"/resource/"+resourceType,
						bytes.NewReader(payload),
					)
					req.Header.Add(trigger.TriggerType, expectedHeaders[trigger.TriggerType])
					req.Header.Add(trigger.TriggerSource, expectedHeaders[trigger.TriggerSource])
					req.Header.Add(trigger.TriggerSourceNamespace, expectedHeaders[trigger.TriggerSourceNamespace])

					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusCreated))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"id":"` + response + `"}`))
				})
			})

			When("request body fails to be read", func() {
				It("returns 500 with error response body", func() {
					req := httptest.NewRequest(
						http.MethodPost,
						"/resource/"+resourceType,
						io.NopCloser(&failReader{}),
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"providerError":"failed to read request body"}`))
				})
			})

			When("handledResource Create fails", func() {
				When("the error is UserError", func() {
					It("returns 400 with error response body", func() {
						response := "failed to create"
						handledResource.On("Create", ignoreCtx, payload, mock.Anything).Return(
							nil,
							&resource.UserError{E: errors.New(response)},
						)

						req := httptest.NewRequest(
							http.MethodPost,
							"/resource/"+resourceType,
							bytes.NewReader(payload),
						)
						rr := httptest.NewRecorder()
						server.Config.Handler.ServeHTTP(rr, req)
						resp := rr.Result()

						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

						body, err := io.ReadAll(resp.Body)

						Expect(err).ToNot(HaveOccurred())
						Expect(string(body)).To(Equal(`{"providerError":"` + response + `"}`))
					})
				})
				When("the error is UnimplementedError", func() {
					It("returns 501 with error response body", func() {
						response := resource.UnimplementedError{
							Method:       "Create",
							ResourceType: resourceType,
						}
						handledResource.On("Create", ignoreCtx, payload, mock.Anything).Return(
							nil,
							&response,
						)

						req := httptest.NewRequest(
							http.MethodPost,
							"/resource/"+resourceType,
							bytes.NewReader(payload),
						)
						rr := httptest.NewRecorder()
						server.Config.Handler.ServeHTTP(rr, req)
						resp := rr.Result()

						Expect(resp.StatusCode).To(Equal(http.StatusNotImplemented))

						body, err := io.ReadAll(resp.Body)

						Expect(err).ToNot(HaveOccurred())
						Expect(string(body)).To(Equal(`{"providerError":"` + response.Error() + `"}`))
					})
				})
				It("returns 500 with error response body", func() {
					response := "failed to create"
					handledResource.On("Create", ignoreCtx, payload, mock.Anything).Return(
						nil,
						errors.New(response),
					)

					req := httptest.NewRequest(
						http.MethodPost,
						"/resource/"+resourceType,
						bytes.NewReader(payload),
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"providerError":"` + response + `"}`))
				})
			})
		})

		Context("/{id} PUT request updateHandler", func() {
			When("succeeds", func() {
				It("returns 200 with valid response body", func() {
					id := "mock-id/bla"
					response := "response-id"
					handledResource.On("Update", ignoreCtx, id, payload, mock.Anything).Return(
						resource.ResponseBody{Id: response},
						nil,
					)
					encodedId := url.PathEscape(id)
					req := httptest.NewRequest(
						http.MethodPut,
						"/resource/"+resourceType+"/"+encodedId,
						bytes.NewReader(payload),
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"id":"` + response + `"}`))
				})
			})

			When("request id fails to decode", func() {
				It("returns 400 with error response body", func() {
					invalidId := "mock-id-broken%"
					req := httptest.NewRequest(
						http.MethodPut,
						"/to-be-overriden",
						bytes.NewReader(payload),
					)
					// cannot set a URL that fails decoding in httptest.NewRequest
					// because it gets validated at construction.
					req.URL.Path = "/resource/" + resourceType + "/" + invalidId
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"providerError":"invalid URL escape \"%\""}`))
				})
			})

			When("request body fails to be read", func() {
				It("returns 500 with error response body", func() {
					req := httptest.NewRequest(
						http.MethodPut,
						"/resource/"+resourceType+"/mock-id",
						io.NopCloser(&failReader{}),
					)
					rr := httptest.NewRecorder()

					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"id":"mock-id","providerError":"failed to read request body"}`))
				})
			})

			When("handledResource Update fails", func() {
				When("the error is UserError", func() {
					It("returns 400 with error response body", func() {
						id := "mock-id"
						response := "failed to update"
						handledResource.On("Update", ignoreCtx, id, payload, mock.Anything).Return(
							nil,
							&resource.UserError{E: errors.New(response)},
						)
						req := httptest.NewRequest(
							http.MethodPut,
							"/resource/"+resourceType+"/"+id,
							bytes.NewReader(payload),
						)
						rr := httptest.NewRecorder()
						server.Config.Handler.ServeHTTP(rr, req)
						resp := rr.Result()

						Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

						body, err := io.ReadAll(resp.Body)

						Expect(err).ToNot(HaveOccurred())
						Expect(string(body)).To(Equal(`{"id":"mock-id","providerError":"` + response + `"}`))
					})
				})
				When("the error is UnimplementedError", func() {
					It("returns 501 with error response body", func() {
						id := "mock-id"
						response := resource.UnimplementedError{
							Method:       "Update",
							ResourceType: resourceType,
						}
						handledResource.On("Update", ignoreCtx, id, payload, mock.Anything).Return(
							nil,
							&response,
						)
						req := httptest.NewRequest(
							http.MethodPut,
							"/resource/"+resourceType+"/"+id,
							bytes.NewReader(payload),
						)
						rr := httptest.NewRecorder()
						server.Config.Handler.ServeHTTP(rr, req)
						resp := rr.Result()

						Expect(resp.StatusCode).To(Equal(http.StatusNotImplemented))

						body, err := io.ReadAll(resp.Body)

						Expect(err).ToNot(HaveOccurred())
						Expect(string(body)).To(Equal(`{"id":"mock-id","providerError":"` + response.Error() + `"}`))
					})
				})
				It("returns 500 with error response body", func() {
					id := "mock-id"
					response := "failed to update"
					handledResource.On("Update", ignoreCtx, id, payload, mock.Anything).Return(
						nil,
						errors.New(response),
					)
					req := httptest.NewRequest(
						http.MethodPut,
						"/resource/"+resourceType+"/"+id,
						bytes.NewReader(payload),
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"id":"mock-id","providerError":"` + response + `"}`))
				})
			})
		})

		Context("/{id} DELETE request deleteHandler", func() {
			When("succeeds", func() {
				It("returns 200 with empty body", func() {
					id := "mock-id/bla"
					encodedId := url.PathEscape(id)
					handledResource.On("Delete", ignoreCtx, id, mock.Anything).Return(nil)

					req := httptest.NewRequest(
						http.MethodDelete,
						"/resource/"+resourceType+"/"+encodedId,
						nil,
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal("{}"))
				})
			})

			When("request id fails to decode", func() {
				It("returns 400 with error response body", func() {
					invalidId := "mock-id-broken%"
					req := httptest.NewRequest(
						http.MethodDelete,
						"/to-be-overriden",
						nil,
					)
					// cannot set a URL that fails decoding in httptest.NewRequest
					// because it gets validated at construction.
					req.URL.Path = "/resource/" + resourceType + "/" + invalidId
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"providerError":"invalid URL escape \"%\""}`))
				})
			})

			When("handledResource Delete fails", func() {
				When("the error is UnimplementedError", func() {
					It("returns 501 with error response body", func() {
						id := "mock-id"
						response := resource.UnimplementedError{
							Method:       "Delete",
							ResourceType: resourceType,
						}
						handledResource.On("Delete", ignoreCtx, id, mock.Anything).Return(&response)
						req := httptest.NewRequest(
							http.MethodDelete,
							"/resource/"+resourceType+"/"+id,
							nil,
						)
						rr := httptest.NewRecorder()
						server.Config.Handler.ServeHTTP(rr, req)
						resp := rr.Result()

						Expect(resp.StatusCode).To(Equal(http.StatusNotImplemented))

						body, err := io.ReadAll(resp.Body)

						Expect(err).ToNot(HaveOccurred())
						Expect(string(body)).To(Equal(`{"id":"mock-id","providerError":"` + response.Error() + `"}`))
					})
				})

				It("returns 500 with error response body", func() {
					id := "mock-id"
					response := "failed to delete"
					handledResource.On("Delete", ignoreCtx, id, mock.Anything).Return(errors.New(response))
					req := httptest.NewRequest(
						http.MethodDelete,
						"/resource/"+resourceType+"/"+id,
						nil,
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"id":"mock-id","providerError":"` + response + `"}`))
				})
			})
		})
	})
})

var _ = Describe("flattenHeaders", func() {
	DescribeTable("flattens and lowercases HTTP headers",
		func(input http.Header, expected map[string]string) {
			result := flattenHeaders(input)
			Expect(result).To(Equal(expected))
		},

		Entry("empty headers",
			http.Header{},
			map[string]string{},
		),

		Entry("single header, single value",
			http.Header{
				"X-Test": []string{"abc"},
			},
			map[string]string{
				"x-test": "abc",
			},
		),

		Entry("single header, multiple values",
			http.Header{
				"X-Test": []string{"a", "b", "c"},
			},
			map[string]string{
				"x-test": "a,b,c",
			},
		),
	)
})
