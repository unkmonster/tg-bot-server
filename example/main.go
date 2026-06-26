package main

import (
	"context"
	"os"
	"strings"

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
	s := tgbotserver.New(api)
	s.OnCommand("echo", func(ctx context.Context, r *tgbotserver.Request) (tgbotapi.Chattable, error) {
		args := r.Args()
		return tgbotapi.NewMessage(0, strings.Join(args, " ")), nil
	})

	if err := s.Start(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	s.Stop(context.TODO())
}
