//go:build unit

package webhook

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
)

func extractHostPort(url string) (string, int) {
	regex := regexp.MustCompile("(?:[^:]+)://([^:]+):([0-9]+)")
	matches := regex.FindStringSubmatch(url)
	port, _ := strconv.Atoi(matches[2])
	return matches[1], port
}

func withHttpWebhook(httpResponseCode int, expectedHeaders http.Header, expectedBody string, f func(upstream HttpWebhook)) func() {
	return func() {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()

			Expect(r.URL.String()).To(Equal("/test-path"))
			for headerName, headerValue := range expectedHeaders {
				Expect(r.Header.Values(headerName)).To(Equal(headerValue))
			}
			content, err := io.ReadAll(r.Body)
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(string(content)).To(Equal(expectedBody))
			w.WriteHeader(httpResponseCode)
		}))
		defer testServer.Close()

		client := testServer.Client()
		testHost, testPort := extractHostPort(testServer.URL)
		underTest := HttpWebhook{Upstream: config.Endpoint{
			Host: testHost,
			Port: testPort,
			Path: "test-path",
		}, Client: client}

		f(underTest)
	}
}

var _ = Context("call", func() {
	var ctx = logr.NewContext(context.Background(), logr.Discard())

	When("called", func() {
		headers := http.Header{"hello": []string{"world", "goodbye"}}
		bodyStr := "hello world"
		rawJson := json.RawMessage(bodyStr)
		eventData := EventData{
			Header: headers,
			Body:   rawJson,
		}

		It("return no error", withHttpWebhook(http.StatusOK, headers, bodyStr, func(underTest HttpWebhook) {
			err := underTest.call(ctx, eventData)
			Expect(err).NotTo(HaveOccurred())
		}))

		It("returns internal server error if upstream fails", withHttpWebhook(http.StatusMethodNotAllowed, headers, bodyStr, func(underTest HttpWebhook) {
			err := underTest.call(ctx, eventData)
			Expect(err).To(HaveOccurred())
		}))

	})
})

var _ = Context("transferHeaders", func() {
	emptyHeaders := make(http.Header)
	headers := map[string][]string{"Foo": {"bar"}, "Baz": {"qux"}}
	DescribeTable("Headers", func(incomingHeaders http.Header, requestHeaders http.Header, expected http.Header) {
		webhook := HttpWebhook{}
		request := http.Request{Header: requestHeaders}
		webhook.transferHeaders(incomingHeaders, &request)
		Expect(request.Header).To(Equal(expected))
	},
		Entry("No incoming headers or existing request headers", emptyHeaders, emptyHeaders, emptyHeaders),
		Entry("No incoming headers are transferred to existing request headers", emptyHeaders, headers, headers),
		Entry("Incoming headers are transferred to empty request headers", headers, emptyHeaders, headers),
		Entry("Incoming headers are transferred to existing request headers", headers, map[string][]string{"Quux": {"corge"}}, map[string][]string{"Quux": {"corge"}, "Foo": {"bar"}, "Baz": {"qux"}}),
		Entry("Incoming headers are combined with existing request headers", map[string][]string{"Foo": {"bar"}}, map[string][]string{"Foo": {"baz"}}, map[string][]string{"Foo": {"baz", "bar"}}),
	)
})
