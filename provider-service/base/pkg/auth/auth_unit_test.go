//go:build unit

package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Unit Suite")
}

var _ = Context("FileTokenSource", func() {
	writeToken := func(content string) string {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "token")
		Expect(os.WriteFile(path, []byte(content), 0600)).To(Succeed())
		return path
	}

	When("the token file exists", func() {
		It("returns the trimmed token", func() {
			path := writeToken("  my-token\n")
			source := NewFileTokenSource(path)
			token, err := source.Token()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("my-token"))
		})

		It("re-reads the file on each call", func() {
			path := writeToken("first")
			source := NewFileTokenSource(path)
			first, err := source.Token()
			Expect(err).NotTo(HaveOccurred())
			Expect(first).To(Equal("first"))

			Expect(os.WriteFile(path, []byte("second"), 0600)).To(Succeed())
			second, err := source.Token()
			Expect(err).NotTo(HaveOccurred())
			Expect(second).To(Equal("second"))
		})
	})

	When("the token file is missing", func() {
		It("returns an error", func() {
			source := NewFileTokenSource(filepath.Join(GinkgoT().TempDir(), "absent"))
			_, err := source.Token()
			Expect(err).To(HaveOccurred())
		})
	})

	When("the token file is empty", func() {
		It("returns an error", func() {
			source := NewFileTokenSource(writeToken("   "))
			_, err := source.Token()
			Expect(err).To(HaveOccurred())
		})
	})
})

// staticTokenSource returns a fixed token, for credential tests.
type staticTokenSource struct{ token string }

func (s staticTokenSource) Token() (string, error) { return s.token, nil }

// errorTokenSource always fails, to exercise error propagation.
type errorTokenSource struct{}

func (errorTokenSource) Token() (string, error) {
	return "", os.ErrNotExist
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
