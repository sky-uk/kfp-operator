package auth

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type BearerCredentials struct {
	Source oauth2.TokenSource
}

func NewBearerCredentials(source oauth2.TokenSource) *BearerCredentials {
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
		"authorization": "Bearer " + token.AccessToken,
	}, nil
}

func (c *BearerCredentials) RequireTransportSecurity() bool {
	return false
}

func BearerDialOption(source oauth2.TokenSource) grpc.DialOption {
	return grpc.WithPerRPCCredentials(NewBearerCredentials(source))
}
