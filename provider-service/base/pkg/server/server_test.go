//go:build unit

package server

import (
	"bytes"
	"errors"
	"testing"

	"io"
	"net/http"
	"net/http/httptest"

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

func (m *MockHandledResource) Update(id string, body []byte) error {
	args := m.Called(id, body)
	return args.Error(0)
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
		handledResource *MockHandledResource
	)

	BeforeEach(func() {
		handledResource = &MockHandledResource{}
		handledResource.On("Type").Return(resourceType)
		server = httptest.NewServer(New([]resource.HttpHandledResource{
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
				It("returns 200 with valid response body", func() {
					response := "mocked-id"
					handledResource.On("Create", mock.Anything).Return(
						resource.ResponseBody{
							Id: response,
						},
						nil,
					)

					req := httptest.NewRequest(
						http.MethodPost,
						"/resource/"+resourceType,
						bytes.NewReader([]byte{}),
					)
					w := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(w, req)
					resp := w.Result()

					Expect(resp.StatusCode).To(Equal(200))
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
					w := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(w, req)
					resp := w.Result()

					Expect(resp.StatusCode).To(Equal(500))
					body, err := io.ReadAll(resp.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal("Failed to read request body\n"))
				})
			})

			When("handledResource Create fails", func() {
				It("returns 500 with error response body", func() {
					response := "failed to create"
					handledResource.On("Create", mock.Anything).Return(
						nil,
						errors.New(response),
					)

					req := httptest.NewRequest(
						http.MethodPost,
						"/resource/"+resourceType,
						bytes.NewReader([]byte{}),
					)
					w := httptest.NewRecorder()
					server.Config.Handler.ServeHTTP(w, req)
					resp := w.Result()

					Expect(resp.StatusCode).To(Equal(500))
					body, err := io.ReadAll(resp.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(ContainSubstring(response))
				})
			})
		})
	})
})
