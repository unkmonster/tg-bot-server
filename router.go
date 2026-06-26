package tgbotserver

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandleRequestFunc func(ctx context.Context, r *Request) (tgbotapi.Chattable, error)

// ConditionHandler if Match return true apply Handler
type ConditionHandler struct {
	Cond    func(r *Request) bool
	Handler HandleRequestFunc
}

type router struct {
	commandHandlers   map[string]HandleRequestFunc
	textHandlers      map[string]HandleRequestFunc
	callbackHandlers  map[string]HandleRequestFunc
	conditionHandlers []ConditionHandler
}

func newRouter() *router {
	return &router{
		commandHandlers:  map[string]HandleRequestFunc{},
		textHandlers:     map[string]HandleRequestFunc{},
		callbackHandlers: map[string]HandleRequestFunc{},
	}
}

func (r *router) OnCommand(cmd string, h HandleRequestFunc) {
	r.commandHandlers[cmd] = h
}

func (r *router) OnText(text string, h HandleRequestFunc) {
	r.textHandlers[text] = h
}

func (r *router) OnCallback(cmd string, h HandleRequestFunc) {
	r.callbackHandlers[cmd] = h
}

func (r *router) AddConditionHandler(h ConditionHandler) {
	r.conditionHandlers = append(r.conditionHandlers, h)
}

type matchResult struct {
	Type     string
	Key      string
	Handlers []HandleRequestFunc
}

func (r *router) match(req *Request) (h HandleRequestFunc) {
	var (
		cmd = req.Cmd()
	)

	switch req.Type() {
	case CmdRequest:
		h = r.commandHandlers[cmd]
	case TextRequest:
		h = r.textHandlers[cmd]
	case CallbackRequest:
		h = r.callbackHandlers[cmd]
	}

	if h == nil {
		for _, ch := range r.conditionHandlers {
			if ch.Cond(req) {
				return ch.Handler
			}
		}
	}

	return
}
