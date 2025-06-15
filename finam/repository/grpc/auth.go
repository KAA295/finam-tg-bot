package grpc

import (
	"context"

	"google.golang.org/grpc"

	auth_service "finam/grpc/tradeapi/v1/auth"
)

type Auth struct {
	client auth_service.AuthServiceClient
}

func NewAuth(cc grpc.ClientConnInterface) *Auth {
	client := auth_service.NewAuthServiceClient(cc)
	return &Auth{client: client}
}

func (a *Auth) Auth(ctx context.Context, secret string) (*auth_service.AuthResponse, error) {
	return a.client.Auth(ctx, &auth_service.AuthRequest{Secret: secret})
}
