package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	grpcRepo "finam/repository/grpc"
	"finam/service"
)

const (
	EndPoint = ""
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env found")
	}
}

func main() {
	tlsConfig := tls.Config{MinVersion: tls.VersionTLS12}
	conn, err := grpc.NewClient("api.finam.ru:443", grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)))
	if err != nil {
		panic(err)
	}
	secret, exists := os.LookupEnv("SECRET")
	if !exists {
		log.Fatal("secret not found")
	}
	ctx := context.Background()

	authRepo := grpcRepo.NewAuth(conn)
	authService := service.NewAuth(authRepo)
	jwt, err := authService.GetToken(context.TODO(), secret)
	if err != nil {
		log.Fatal("couldn't get token")
	}

	md := metadata.New(map[string]string{
		"Authorization": jwt,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	accountID, exists := os.LookupEnv("ACCOUNT_ID")
	if !exists {
		log.Fatal("account_id not found")
	}

	accountRepo := grpcRepo.NewAccount(conn, accountID)
	accountService := service.NewAccount(accountRepo)
	equity, err := accountService.GetEquity(ctx)
	if err != nil {
		log.Fatalf("can't get equity: %v", err)
	}
	fmt.Println(equity)
}
