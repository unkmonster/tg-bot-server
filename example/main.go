package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgbotserver "github.com/unkmonster/tg-bot-server"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	s := tgbotserver.New(
		api,
		tgbotserver.WithMiddleware(
			tgbotserver.Logging(log.DefaultLogger),
			tgbotserver.Panic(),
		),
	)

	s.OnCommand("echo", func(ctx context.Context, r *tgbotserver.Request) (tgbotapi.Chattable, error) {
		args := r.Args()
		return tgbotapi.NewMessage(0, strings.Join(args, " ")), nil
	})
	s.OnCommand("error", func(ctx context.Context, r *tgbotserver.Request) (tgbotapi.Chattable, error) {
		return nil, errors.New(500, "TEST_ERROR", "手动触发")
	})
	s.OnCommand("panic", func(ctx context.Context, r *tgbotserver.Request) (tgbotapi.Chattable, error) {
		panic("test")
	})

	if err := s.Start(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	s.Stop(context.TODO())
}
