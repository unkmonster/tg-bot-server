package tgbotserver

import "context"

type Session struct {
	SenderID int64
	ChatID   int64
}

type sessionKey struct{}

func NewContext(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

func FromContext(ctx context.Context) (s *Session, ok bool) {
	s, ok = ctx.Value(sessionKey{}).(*Session)
	return
}
