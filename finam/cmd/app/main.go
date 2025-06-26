package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"finam/config"
	grpcRepo "finam/repository/grpc"
	"finam/repository/http"
	redisRepo "finam/repository/redis"
	"finam/service"
)

const (
	FinamEndPoint = "api.finam.ru:443"
	TGEndPoint    = "api.telegram.org"
)

// func init() {
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatal("No .env found")
// 	}
// }

func main() {
	serviceConfig := `{
        "methodConfig": [{
            "retryPolicy": {
                "maxAttempts": 4,
                "initialBackoff": "0.1s",
                "maxBackoff": "5s",
                "backoffMultiplier": 2,
                "retryableStatusCodes": ["UNAVAILABLE"]
            }
        }]
    }`
	fl := config.ParseFlags()
	var cfg config.Config

	config.MustLoad(fl.ConfigPath, &cfg)
	tlsConfig := tls.Config{MinVersion: tls.VersionTLS12}
	conn, err := grpc.NewClient(FinamEndPoint, grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)), grpc.WithDefaultServiceConfig(serviceConfig))
	if err != nil {
		panic(err)
	}
	secret, exists := os.LookupEnv("SECRET")
	if !exists {
		log.Fatal("SECRET not found")
	}
	ctx := context.Background()

	authRepo := grpcRepo.NewAuth(conn)
	authService := service.NewAuth(authRepo)

	accountID, exists := os.LookupEnv("ACCOUNT_ID")
	if !exists {
		log.Fatal("ACCOUNT_ID not found")
	}

	accountRepo := grpcRepo.NewAccount(conn, accountID)
	accountService := service.NewAccount(accountRepo, authService)

	tgToken, exists := os.LookupEnv("TG_TOKEN")
	if !exists {
		log.Fatal("TG_TOKEN not found")
	}

	tgUser, exists := os.LookupEnv("TG_USER")
	if !exists {
		log.Fatal("TG_USER not found")
	}
	tgUserID, err := strconv.Atoi(tgUser)
	if err != nil {
		log.Fatal("tg user must be int")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
		DB:   cfg.Redis.DB,
	})

	cacheRepo := redisRepo.NewCache(rdb)

	tgRepo := http.NewTG(TGEndPoint, tgToken)
	tgService := service.NewTg(tgUserID, tgRepo, cacheRepo, accountService, authService, secret, cfg.Notification.StartHour, cfg.Notification.StartMinute, time.Duration(cfg.Notification.NSInterval), cfg.Chart.NLatest)
	err = tgService.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
