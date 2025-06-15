package grpc

import (
	"context"

	"google.golang.org/grpc"

	accounts_service "finam/grpc/tradeapi/v1/accounts"
)

type Account struct {
	accountID string
	client    accounts_service.AccountsServiceClient
}

func NewAccount(cc grpc.ClientConnInterface, accountID string) *Account {
	client := accounts_service.NewAccountsServiceClient(cc)
	return &Account{accountID: accountID, client: client}
}

func (a *Account) GetAccount(ctx context.Context) (*accounts_service.GetAccountResponse, error) {
	return a.client.GetAccount(ctx, &accounts_service.GetAccountRequest{AccountId: a.accountID})
}
