//go:build unit

package auth

import (
	"context"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Unit Suite")
}

type staticTokenSource struct{ token string }

func (s staticTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: s.token}, nil
}

type errorTokenSource struct{}

func (errorTokenSource) Token() (*oauth2.Token, error) {
	return nil, os.ErrNotExist
}

var _ = Context("BearerCredentials", func() {
	It("returns the authorization header", func() {
		cred := NewBearerCredentials(staticTokenSource{token: "abc"})
		md, err := cred.GetRequestMetadata(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(md).To(HaveKeyWithValue("authorization", "Bearer abc"))
	})

	It("propagates token errors", func() {
		cred := NewBearerCredentials(errorTokenSource{})
		_, err := cred.GetRequestMetadata(context.Background())
		Expect(err).To(HaveOccurred())
	})

	It("does not require transport security", func() {
		cred := NewBearerCredentials(staticTokenSource{token: "abc"})
		Expect(cred.RequireTransportSecurity()).To(BeFalse())
	})
})

var _ = Context("BearerAuthInfoWriter", func() {
	It("sets the authorization header", func() {
		writer := BearerAuthInfoWriter(staticTokenSource{token: "abc"})
		req := &runtime.TestClientRequest{}
		Expect(writer.AuthenticateRequest(req, strfmt.Default)).To(Succeed())
		Expect(req.Headers.Get("Authorization")).To(Equal("Bearer abc"))
	})

	It("propagates token errors", func() {
		writer := BearerAuthInfoWriter(errorTokenSource{})
		req := &runtime.TestClientRequest{}
		Expect(writer.AuthenticateRequest(req, strfmt.Default)).To(HaveOccurred())
	})
})
