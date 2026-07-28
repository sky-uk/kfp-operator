package auth

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// BearerAuthInfoWriter returns a runtime.ClientAuthInfoWriter that sets the
// "Authorization: Bearer <token>" header on go-openapi REST requests. The token
// is read from source on each request so rotated tokens are picked up.
func BearerAuthInfoWriter(source TokenSource) runtime.ClientAuthInfoWriter {
	return runtime.ClientAuthInfoWriterFunc(
		func(req runtime.ClientRequest, _ strfmt.Registry) error {
			token, err := source.Token()
			if err != nil {
				return err
			}
			return req.SetHeaderParam("Authorization", "Bearer "+token)
		},
	)
}
