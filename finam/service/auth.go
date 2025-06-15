package service

import (
	"context"

	auth_service "finam/grpc/tradeapi/v1/auth"
)

type AuthRepo interface {
	Auth(ctx context.Context, secret string) (*auth_service.AuthResponse, error)
}

type Auth struct {
	authRepo AuthRepo
}

func NewAuth(authRepo AuthRepo) *Auth {
	return &Auth{authRepo: authRepo}
}

func (a *Auth) GetToken(ctx context.Context, secret string) (string, error) {
	resp, err := a.authRepo.Auth(ctx, secret)
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}
