package util

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

// JWT implements the gRPC credentials.PerRPCCredentials interface.
// It can be used to inject JWT token into gRPC metadata headers.
type jwtCreds struct {
	token string
}

func NewJwtCreds(token string) credentials.PerRPCCredentials {
	return jwtCreds{token}
}

func (jwt jwtCreds) GetRequestMetadata(
	ctx context.Context,
	uri ...string,
) (map[string]string, error) {
	return map[string]string{
		"authorization": jwt.token,
	}, nil
}

func (j jwtCreds) RequireTransportSecurity() bool {
	return true
}
