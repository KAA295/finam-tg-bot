package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"

	"finam/domain"
)

type TgRepo interface {
	SendMessage(chatID int, text string) error
}

type CacheRepo interface {
	Set(ctx context.Context, values ...string)
	GetLatest(ctx context.Context) (string, error)
}

type AccountService interface {
	GetEquity(ctx context.Context, token domain.Token) (string, error)
}

type AuthService interface {
	UpdateToken(ctx context.Context, secret string, token domain.Token) (domain.Token, error)
}

type Tg struct {
	userID         int
	tgRepo         TgRepo
	cacheRepo      CacheRepo
	accountService AccountService
	authService    AuthService
	secret         string
	accessToken    domain.Token
	startHour      int
	startMinute    int
	interval       time.Duration
}

func NewTg(userID int, tgRepo TgRepo, cacheRepo CacheRepo, accountService AccountService, authService AuthService, secret string, startHour int, startMinute int, interval time.Duration) *Tg {
	return &Tg{userID: userID, tgRepo: tgRepo, cacheRepo: cacheRepo, accountService: accountService, authService: authService, secret: secret, accessToken: domain.Token{}, startHour: startHour, startMinute: startMinute, interval: interval}
}

func (tg *Tg) Start(ctx context.Context) error {
	for {
		time.Sleep(tg.duration())
		token, err := tg.authService.UpdateToken(ctx, tg.secret, tg.accessToken)
		tg.accessToken = token
		if err != nil {
			return err
		}
		md := metadata.New(map[string]string{
			"Authorization": token.Token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
		equity, err := tg.accountService.GetEquity(ctx, tg.accessToken)
		if err != nil {
			return err
		}

		latest, err := tg.cacheRepo.GetLatest(ctx)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				latest = equity
			} else {
				return fmt.Errorf("couldn't get latest from redis: %v", err)
			}
		}

		intEquity, err := strconv.ParseFloat(equity, 64)
		if err != nil {
			return fmt.Errorf("couldn't convert equity to integer: %v", err)
		}
		intLatest, err := strconv.ParseFloat(latest, 64)
		if err != nil {
			return fmt.Errorf("couldn't convert latest to integer: %v", err)
		}

		tg.tgRepo.SendMessage(tg.userID, fmt.Sprintf("Текущий баланс: %.2f\nРазница с предыдущим: %.2f\nПроцентное соотношение: %.3f%%", intEquity, intEquity-intLatest, (intEquity-intLatest)/(intLatest/100)))
		tg.cacheRepo.Set(ctx, equity)
	}
}

func (tg *Tg) duration() time.Duration {
	now := time.Now()
	firstRun := time.Date(now.Year(), now.Month(), now.Day(), tg.startHour, tg.startMinute, 0, 0, now.Location())
	elapsed := now.Sub(firstRun)
	next := firstRun.Add(((elapsed / tg.interval) + 1) * tg.interval)

	return next.Sub(now)
}
