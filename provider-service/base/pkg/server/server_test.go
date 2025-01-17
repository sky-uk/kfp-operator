//go:build unit

package server

import (
	"bytes"
	"errors"
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

func (m *MockHandledResource) Create(body []byte) (resource.ResponseBody, error) {
	args := m.Called(body)
	var response resource.ResponseBody
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(resource.ResponseBody)
	}
	return response, args.Error(1)
}

func (m *MockHandledResource) Update(
	id string,
	body []byte,
) (resource.ResponseBody, error) {
	args := m.Called(id, body)
	var response resource.ResponseBody
	if arg0 := args.Get(0); arg0 != nil {
		response = arg0.(resource.ResponseBody)
	}
	return response, args.Error(1)
}

func (m *MockHandledResource) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type failReader struct{}

func (f *failReader) Read(p []byte) (int, error) {
	return 0, errors.New("simulated read error")
}

var _ = Describe("Http Server Endpoints", func() {
	var (
		server          *httptest.Server
		resourceType    string = "mock-resource"
		payload         []byte = []byte(`{"name": "test"}`)
		handledResource *MockHandledResource
	)

	BeforeEach(func() {
		handledResource = &MockHandledResource{}
		handledResource.On("Type").Return(resourceType)
		server = httptest.NewServer(newHandler([]resource.HttpHandledResource{
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
					handledResource.On("Create", payload).Return(
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
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusCreated))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(response))
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
					Expect(string(body)).To(ContainSubstring("Failed to read request body"))
				})
			})

			When("handledResource Create fails", func() {
				When("the error is UserError", func() {
					It("returns 400 with error response body", func() {
						response := "failed to create"
						handledResource.On("Create", payload).Return(
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
						Expect(string(body)).To(ContainSubstring(response))
					})
				})
				It("returns 500 with error response body", func() {
					response := "failed to create"
					handledResource.On("Create", payload).Return(
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
					Expect(string(body)).To(ContainSubstring(response))
				})
			})
		})

		Context("/{id} PUT request updateHandler", func() {
			When("succeeds", func() {
				It("returns 200 with valid response body", func() {
					id := "mock-id/bla"
					response := "response-id"
					handledResource.On("Update", id, payload).Return(
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
					Expect(string(body)).To(ContainSubstring(response))
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
					Expect(body).To(ContainSubstring(`invalid URL escape "%"`))
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
					Expect(string(body)).To(ContainSubstring("Failed to read request body"))
				})
			})

			When("handledResource Update fails", func() {
				When("the error is UserError", func() {
					It("returns 400 with error response body", func() {
						id := "mock-id"
						response := "failed to update"
						handledResource.On("Update", id, payload).Return(
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
						Expect(string(body)).To(ContainSubstring(response))
					})
				})
				It("returns 500 with error response body", func() {
					id := "mock-id"
					response := "failed to update"
					handledResource.On("Update", id, payload).Return(
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
					Expect(string(body)).To(ContainSubstring(response))
				})
			})
		})

		Context("/{id} DELETE request deleteHandler", func() {
			When("succeeds", func() {
				It("returns 204", func() {
					id := "mock-id/bla"
					encodedId := url.PathEscape(id)
					handledResource.On("Delete", id).Return(nil)

					req := httptest.NewRequest(
						http.MethodDelete,
						"/resource/"+resourceType+"/"+encodedId,
						nil,
					)
					rr := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(rr, req)
					resp := rr.Result()

					Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

					body, err := io.ReadAll(resp.Body)

					Expect(err).ToNot(HaveOccurred())
					Expect(body).To(BeEmpty())
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
					Expect(body).To(ContainSubstring(`invalid URL escape "%"`))
				})
			})

			When("handledResource Delete fails", func() {
				It("returns 500 with error response body", func() {
					id := "mock-id"
					response := "failed to delete"
					handledResource.On("Delete", id).Return(errors.New(response))
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
					Expect(string(body)).To(ContainSubstring(response))
				})
			})
		})
	})
})
