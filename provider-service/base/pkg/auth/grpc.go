package auth

import (
	"context"

	"google.golang.org/grpc"
)

// BearerCredentials implements google.golang.org/grpc/credentials.PerRPCCredentials
// by attaching an "authorization: Bearer <token>" header to every gRPC call.
type BearerCredentials struct {
	Source TokenSource
}

func NewBearerCredentials(source TokenSource) *BearerCredentials {
	return &BearerCredentials{Source: source}
}

func (c *BearerCredentials) GetRequestMetadata(
	_ context.Context,
	_ ...string,
) (map[string]string, error) {
	token, err := c.Source.Token()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

// RequireTransportSecurity returns false so the credential can be sent over a
// plaintext in-cluster connection.
func (c *BearerCredentials) RequireTransportSecurity() bool {
	return false
}

// BearerDialOption returns a gRPC dial option that attaches a bearer token from
// source to every outgoing call.
func BearerDialOption(source TokenSource) grpc.DialOption {
	return grpc.WithPerRPCCredentials(NewBearerCredentials(source))
}
