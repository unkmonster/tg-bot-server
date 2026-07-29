package tgbotserver

import (
	"context"
)

type HandleFunc func(ctx context.Context, r *Request) (*Reply, error)

// ConditionHandler if Match return true apply Handler
type ConditionHandler struct {
	Cond    func(r *Request) bool
	Handler HandleFunc
}

type router struct {
	commandHandlers   map[string]HandleFunc
	textHandlers      map[string]HandleFunc
	callbackHandlers  map[string]HandleFunc
	conditionHandlers []ConditionHandler
}

func newRouter() *router {
	return &router{
		commandHandlers:  map[string]HandleFunc{},
		textHandlers:     map[string]HandleFunc{},
		callbackHandlers: map[string]HandleFunc{},
	}
}

func (r *router) OnCommand(cmd string, h HandleFunc) {
	r.commandHandlers[cmd] = h
}

func (r *router) OnText(text string, h HandleFunc) {
	r.textHandlers[text] = h
}

func (r *router) OnCallback(cmd string, h HandleFunc) {
	r.callbackHandlers[cmd] = h
}

func (r *router) AddConditionHandler(h ConditionHandler) {
	r.conditionHandlers = append(r.conditionHandlers, h)
}

type matchResult struct {
	Type     string
	Key      string
	Handlers []HandleFunc
}

func (r *router) match(req *Request) (h HandleFunc) {
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
