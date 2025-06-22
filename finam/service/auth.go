package service

import (
	"context"
	"time"

	"finam/domain"
	auth_service "finam/grpc/tradeapi/v1/auth"
)

type AuthRepo interface {
	Auth(ctx context.Context, secret string) (*auth_service.AuthResponse, error)
	TokenDetails(ctx context.Context, token string) (*auth_service.TokenDetailsResponse, error)
}

type Auth struct {
	authRepo AuthRepo
}

func NewAuth(authRepo AuthRepo) *Auth {
	return &Auth{authRepo: authRepo}
}

func (a *Auth) GetToken(ctx context.Context, secret string) (domain.Token, error) {
	resp, err := a.authRepo.Auth(ctx, secret)
	if err != nil {
		return domain.Token{}, err
	}
	details, err := a.authRepo.TokenDetails(ctx, resp.Token)
	return domain.Token{Token: resp.Token, Exp: details.ExpiresAt.AsTime()}, nil
}

func (a *Auth) UpdateToken(ctx context.Context, secret string, token domain.Token) (domain.Token, error) {
	if token.Token == "" || time.Now().After(token.Exp) {
		return a.GetToken(ctx, secret)
	}
	return token, nil
}
