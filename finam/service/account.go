package service

import (
	"context"

	"finam/domain"
	accounts_service "finam/grpc/tradeapi/v1/accounts"
)

type AccountRepo interface {
	GetAccount(ctx context.Context) (*accounts_service.GetAccountResponse, error)
}

type Account struct {
	accountRepo AccountRepo
}

func NewAccount(accountRepo AccountRepo, authService AuthService) *Account {
	return &Account{accountRepo: accountRepo}
}

func (a *Account) GetEquity(ctx context.Context, token domain.Token) (string, error) {
	resp, err := a.accountRepo.GetAccount(ctx)
	return resp.GetEquity().Value, err
}
