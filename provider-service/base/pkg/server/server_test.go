//go:build unit

package server

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"net/http/httptest"
)

func TestServerUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Unit Test")
}

var _ = Describe("Http Server Endpoints", Ordered, func() {
	var testServer *httptest.Server

	BeforeAll(func() {
		testServer = httptest.NewServer(New())
	})

	AfterAll(func() {
		testServer.Close()
	})

	Context("/readyz", func() {
		When("called", func() {
			It("should be OK", func() {
				resp, err := http.Get(testServer.URL + "/readyz")
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
				resp, err := http.Get(testServer.URL + "/livez")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(body)).To(Equal("Application is live."))
			})
		})
	})
})
