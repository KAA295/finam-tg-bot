package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"google.golang.org/grpc/metadata"

	"finam/domain"
)

type TgRepo interface {
	SendMessage(chatID int, text string) error
	SendPhoto(chatID int, caption string, filename string) error
}

type CacheService interface {
	Set(ctx context.Context, values ...string)
	GetLatest(ctx context.Context) (string, error)
	GetNLatest(ctx context.Context, n int) ([]string, error)
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
	cacheService   CacheService
	accountService AccountService
	authService    AuthService
	secret         string
	accessToken    domain.Token
	startHour      int
	startMinute    int
	interval       time.Duration
	nLatest        int
}

func NewTg(userID int, tgRepo TgRepo, cacheService CacheService, accountService AccountService, authService AuthService, secret string, startHour int, startMinute int, interval time.Duration, nLatest int) *Tg {
	return &Tg{userID: userID, tgRepo: tgRepo, cacheService: cacheService, accountService: accountService, authService: authService, secret: secret, accessToken: domain.Token{}, startHour: startHour, startMinute: startMinute, interval: interval, nLatest: nLatest}
}

func (tg *Tg) Start(ctx context.Context) error {
	for {
		time.Sleep(tg.duration())
		token, err := tg.authService.UpdateToken(ctx, tg.secret, tg.accessToken)
		tg.accessToken = token
		if err != nil {
			tg.handleError(ctx, "UpdateToken", err)
			continue
		}
		md := metadata.New(map[string]string{
			"Authorization": token.Token,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
		equity, err := tg.accountService.GetEquity(ctx, tg.accessToken)
		if err != nil {
			tg.handleError(ctx, "GetEquity", err)
			continue
		}

		latest, err := tg.cacheService.GetLatest(ctx)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				latest = equity
			} else {
				tg.handleError(ctx, "GetLatest", err)
				continue
			}
		}

		intEquity, err := strconv.ParseFloat(equity, 64)
		if err != nil {
			tg.handleError(ctx, "ConvertEquity", err)
			continue
		}
		intLatest, err := strconv.ParseFloat(latest, 64)
		if err != nil {
			tg.handleError(ctx, "ConvertLatest", err)
			continue
		}
		filename, err := tg.getCharts(ctx, tg.nLatest)
		if err != nil {
			tg.handleError(ctx, "GetChart", err)
			continue
		}

		err = tg.tgRepo.SendPhoto(tg.userID, fmt.Sprintf("Текущий баланс: %.2f\nРазница с предыдущим: %.2f\nПроцентное соотношение: %.2f%%", intEquity, intEquity-intLatest, (intEquity-intLatest)/(intLatest/99)), filename)
		if err != nil {
			tg.handleError(ctx, "SendPhoto", err)
		}

		tg.cacheService.Set(ctx, equity)
	}
}

func (tg *Tg) duration() time.Duration {
	now := time.Now()
	firstRun := time.Date(now.Year(), now.Month(), now.Day(), tg.startHour, tg.startMinute, 0, 0, now.Location())
	elapsed := now.Sub(firstRun)
	next := firstRun.Add(((elapsed / tg.interval) + 1) * tg.interval)
	fmt.Println(next.Sub(now))
	fmt.Println(next.Sub(now).Minutes())
	return next.Sub(now)
}

func (tg *Tg) getCharts(ctx context.Context, n int) (string, error) {
	p := plot.New()
	p.Title.Text = "График баланса"
	p.X.Label.Text = ""
	p.Y.Label.Text = "Баланс"
	filename := "balance.jpg"
	data, err := tg.cacheService.GetNLatest(ctx, n)
	if err != nil && err != redis.Nil {
		return "", err
	}

	pts := make(plotter.XYs, len(data))
	for i, v := range data {
		pts[i].X = float64(i + 1)
		y, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", err
		}
		pts[i].Y = y
	}
	line, err := plotter.NewLine(pts)
	if err != nil {
		return "", err
	}
	p.Add(line)

	err = p.Save(8*vg.Inch, 4*vg.Inch, filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func (tg *Tg) handleError(ctx context.Context, stage string, err error) {
	log.Printf("error: %w on stage: %v", err, stage)

	go func() {
		tg.tgRepo.SendMessage(tg.userID, "Произошла ошибка")
	}()
}
