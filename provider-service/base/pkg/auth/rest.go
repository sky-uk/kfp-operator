package auth

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"golang.org/x/oauth2"
)

// BearerAuthInfoWriter returns a runtime.ClientAuthInfoWriter that adds an
// "Authorization: Bearer <token>" header to go-openapi REST requests.
func BearerAuthInfoWriter(source oauth2.TokenSource) runtime.ClientAuthInfoWriter {
	return runtime.ClientAuthInfoWriterFunc(
		func(req runtime.ClientRequest, _ strfmt.Registry) error {
			token, err := source.Token()
			if err != nil {
				return err
			}
			return req.SetHeaderParam("Authorization", "Bearer "+token.AccessToken)
		},
	)
}
