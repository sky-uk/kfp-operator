//go:build unit

package server

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VAI Server Unit Test")
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
		When("foo", func() {
			It("should be OK", func() {
				resp, err := http.Get(testServer.URL + "/readyz")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
