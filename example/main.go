package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
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

	s.OnCommand("echo", func(ctx context.Context, r *tgbotserver.Request) (*tgbotserver.Reply, error) {
		args := r.Args()
		text := strings.Join(args, " ")

		if text == "" {
			return nil, errors.BadRequest("EMPTY_ARGS", "参数不可为空")
		}
		return &tgbotserver.Reply{
			Messages: []tgbotapi.Chattable{lo.ToPtr(tgbotapi.NewMessage(0, text))},
		}, nil
	})
	s.OnCommand("error", func(ctx context.Context, r *tgbotserver.Request) (*tgbotserver.Reply, error) {
		return nil, errors.New(500, "TEST_ERROR", "手动触发")
	})
	s.OnCommand("panic", func(ctx context.Context, r *tgbotserver.Request) (*tgbotserver.Reply, error) {
		panic("test")
	})

	if err := s.Start(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	s.Stop(context.TODO())
}
