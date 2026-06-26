package tgbotserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	DefaultSep = " "

	CmdRequest      = "cmd"
	TextRequest     = "text"
	CallbackRequest = "callback"
	AnyRequest      = "any"
)

type Request struct {
	Update tgbotapi.Update

	typ  string // request type
	cmd  string
	args []string // parsed arguments with cmd
}

func NewRequest(update tgbotapi.Update) *Request {
	rv := &Request{
		Update: update,
	}

	// if c := update.FromChat(); c != nil {
	// 	rv.chatID = c.ID
	// }
	// if s := update.SentFrom(); s != nil {
	// 	rv.senderID = s.ID
	// }

	if update.Message != nil {
		if update.Message.IsCommand() {
			rv.typ = CmdRequest
			rv.args = strings.Split(update.Message.CommandArguments(), DefaultSep)
			rv.cmd = update.Message.Command()
			return rv
		}

		fullArgs := strings.Split(update.Message.Text, DefaultSep)
		rv.typ = TextRequest
		if len(fullArgs) > 0 {
			rv.cmd = fullArgs[0]
		}
		if len(fullArgs) > 1 {
			rv.args = fullArgs[1:]
		}
		return rv
	}

	if update.CallbackQuery != nil {
		rv.typ = CallbackRequest
		fullArgs := strings.Split(update.CallbackData(), DefaultSep)
		if len(fullArgs) > 0 {
			rv.cmd = fullArgs[0]
		}
		if len(fullArgs) > 1 {
			rv.args = fullArgs[1:]
		}
		return rv
	}
	return rv
}

// Cmd returns request cmd
func (r *Request) Cmd() string {
	return r.cmd
}

// Type returns request type
func (r *Request) Type() string {
	return r.typ
}

// Args returns deep copy of args
func (r *Request) Args() []string {
	rv := make([]string, len(r.args))
	copy(rv, r.args)
	return rv
}

type Option func(b *Bot)

func WithUpdateConfig(c *tgbotapi.UpdateConfig) Option {
	return func(b *Bot) {
		b.updateConfig = c
	}
}

func buildHelper(logger log.Logger) *log.Helper {
	return log.NewHelper(log.With(logger, "component", "bot"))
}

func WithLogger(logger log.Logger) Option {
	return func(b *Bot) {
		b.log = buildHelper(logger)
	}
}

type Bot struct {
	*router
	log *log.Helper

	api          *tgbotapi.BotAPI
	updateConfig *tgbotapi.UpdateConfig

	cancel context.CancelFunc
	done   chan struct{}
}

func New(
	api *tgbotapi.BotAPI,
	opts ...Option,
) *Bot {
	b := &Bot{
		router: newRouter(),
		log:    buildHelper(log.DefaultLogger),
		api:    api,
		done:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Bot) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	go b.worker(ctx)
	b.log.WithContext(ctx).Infof("[Bot] worker started")
	return nil
}

func (b *Bot) Stop(ctx context.Context) error {
	b.log.WithContext(ctx).Infof("[Bot] worker stopping")
	b.cancel()
	<-b.done
	return nil
}

func (b *Bot) worker(ctx context.Context) {
	defer close(b.done)
	var cfg tgbotapi.UpdateConfig
	if b.updateConfig != nil {
		cfg = *b.updateConfig
	} else {
		// new default update config
		cfg = tgbotapi.NewUpdate(0)
		cfg.Timeout = 60
	}

	updates := b.api.GetUpdatesChan(cfg)
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, up tgbotapi.Update) {
	r := NewRequest(up)

	s := newSession(up)
	ctx = NewContext(ctx, s)

	h := b.router.match(r)
	if h == nil {
		b.log.WithContext(ctx).Debugw(
			"event", "no any handler matched",
			"request.cmd", r.Cmd(),
			"request.type", r.Type(),
		)
		return
	}

	reply, err := h(ctx, r)
	if err != nil {
		reply = formatError(err)
	}
	reply = setReplyDefaults(reply, up)

	_, err = b.api.Request(reply)
	if err != nil {
		b.log.WithContext(ctx).Errorw(
			"msg", "failed to send reply",
			"reason", err,
			"update.id", up.UpdateID,
			"session.chat_id", s.ChatID,
			"session.sender_id", s.SenderID,
			"request.cmd", r.Cmd(),
			"request.type", r.Type(),
		)
		return
	}
}

func newSession(up tgbotapi.Update) *Session {
	s := &Session{}
	if user := up.SentFrom(); user != nil {
		s.SenderID = user.ID
	}

	if chat := up.FromChat(); chat != nil {
		s.ChatID = chat.ID
	}
	return s
}

func newContext(ctx context.Context, up tgbotapi.Update) context.Context {
	s := newSession(up)
	return NewContext(ctx, s)
}

func formatError(err error) tgbotapi.MessageConfig {
	se := errors.FromError(err)
	if se.Reason == "" {
		se.Message = "内部错误"
	}
	text := se.Message

	if data, err := json.Marshal(se); err == nil {
		text = string(data)
	}
	msg := tgbotapi.NewMessage(0, fmt.Sprintf("<pre>%s</pre>", text))
	msg.ParseMode = "html"
	return msg
}

func setMessageReplyDefaults(v *tgbotapi.MessageConfig, up tgbotapi.Update) {
	if v == nil {
		return
	}

	var (
		chatId   int64
		senderId int64
	)

	if up.FromChat() != nil {
		chatId = up.FromChat().ID
	}
	if up.SentFrom() != nil {
		senderId = up.SentFrom().ID
	}

	if v.ChatID == 0 {
		v.ChatID = chatId
	}
	// group mode
	if v.ReplyToMessageID == 0 && chatId != senderId {
		v.ReplyToMessageID = int(senderId)
	}

	if v.ParseMode == "" {
		v.ParseMode = "html"
	}
}

func setCallbackReplyDefaults(v *tgbotapi.CallbackConfig, up tgbotapi.Update) {
	var (
		queryId string
	)

	if up.CallbackQuery != nil {
		queryId = up.CallbackQuery.ID
	}

	if v.CallbackQueryID == "" {
		v.CallbackQueryID = queryId
	}
}

func setReplyDefaults(reply tgbotapi.Chattable, up tgbotapi.Update) tgbotapi.Chattable {
	if reply == nil {
		return nil
	}

	switch v := reply.(type) {
	case tgbotapi.MessageConfig:
		setMessageReplyDefaults(&v, up)
		return v
	case *tgbotapi.MessageConfig:
		setMessageReplyDefaults(v, up)
		return v
	case tgbotapi.CallbackConfig:
		setCallbackReplyDefaults(&v, up)
		return v
	case *tgbotapi.CallbackConfig:
		setCallbackReplyDefaults(v, up)
		return v
	}
	return reply
}
