package tgbotserver

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Middleware func(next HandleFunc) HandleFunc

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next HandleFunc) HandleFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func Logging(logger log.Logger) Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, r *Request) (*Reply, error) {
			var (
				start          = time.Now()
				level          = log.LevelInfo
				chatID         int64
				chatUsername   string
				senderID       int64
				senderUsername string
				//callbackQueryID string
			)

			if c := r.Update.FromChat(); c != nil {
				chatID = c.ID
				chatUsername = c.UserName
			}
			if s := r.Update.SentFrom(); s != nil {
				senderID = s.ID
				senderUsername = s.UserName
			}

			rv, err := next(ctx, r)
			if err != nil {
				level = log.LevelError
			}

			kvs := []any{
				"component", "middleware/logging",
				"server", "tg_bot",
				"type", r.Type(),
				"cmd", r.Cmd(),
				"args", fmt.Sprintf("%+v", r.Args()),
				"update.id", r.Update.UpdateID,
				"chat.id", chatID,
				"chat.username", chatUsername,
				"sender.id", senderID,
				"sender.username", senderUsername,
				"latency", time.Since(start),
			}
			if err != nil {
				kvs = append(kvs, "error", err)
			}
			kvs = append(kvs, "msg", "server request")

			logger.Log(
				level,
				kvs...,
			)

			return rv, err
		}
	}
}

func Panic() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, r *Request) (reply *Reply, err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}
				err = fmt.Errorf("recover: %v", r)
			}()
			reply, err = next(ctx, r)
			return
		}
	}
}
