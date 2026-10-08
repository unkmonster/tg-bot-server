# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`tgbotserver` (module `github.com/unkmonster/tg-bot-server`) is a small Go library/framework that wraps
[go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) to add command/text/callback
routing and middleware to Telegram bots. It is a library, not a standalone service — `example/main.go` is a
sample consumer, not the product.

The module depends on a fork of telegram-bot-api, pinned via a `replace` directive in `go.mod`:
`github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/OvyFlash/telegram-bot-api`. Import the
package under its original `tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"` path; the fork
is resolved transparently.

## Commands

- Build: `go build ./...`
- Vet: `go vet ./...`
- Run the example bot: `BOT_TOKEN=<token> go run ./example`
- There are no tests in this repo currently.

## Architecture

The library is organized around four cooperating pieces, all in package `tgbotserver` at the repo root:

- **`bot.go` — `Bot`**: the entry point. `New(api, opts...)` builds a `*Bot` embedding a `*router`.
  `Start`/`Stop` run a goroutine (`worker`) that reads `tgbotapi.Update`s from `api.GetUpdatesChan` and
  dispatches each one to `handleUpdate` in its own goroutine. `handleUpdate` wraps the update in a
  `*Request`, builds a `*Session` and injects it into `context.Context`, looks up a handler via
  `router.match`, runs it through the configured middleware chain, and sends the resulting `*Reply`
  messages (or a formatted error) back via the Telegram API. `Request` normalizes three update shapes
  (command, plain text, callback query) into a common `(typ, cmd, args)` triple using `DefaultSep` (`" "`)
  to split arguments.
- **`router.go` — `router`**: holds three maps (`commandHandlers`, `textHandlers`, `callbackHandlers`)
  keyed by `cmd`, registered via `OnCommand`/`OnText`/`OnCallback`, plus a list of `ConditionHandler`s
  (arbitrary `Cond(*Request) bool` predicates) added via `AddConditionHandler` and checked in order as a
  fallback when no exact key match is found. `Bot` embeds `*router`, so these methods are called directly
  on a `*Bot`.
- **`middleware.go` — `Middleware`**: standard `func(HandleFunc) HandleFunc` chain, composed with
  `MiddlewareChain` (outermost-first) and installed on the bot via `WithMiddleware`. Two middlewares ship
  built-in: `Logging(logger)` (structured request/latency/error logging) and `Panic()` (recovers panics in
  a handler and turns them into an `error`). Middleware wraps the handler returned by `router.match`, not
  the router lookup itself.
- **`session.go` — `Session`**: a minimal `{SenderID, ChatID}` struct stashed in the request `context.Context`
  by `handleUpdate` via `NewContext`/`FromContext`, so handlers (or nested middleware) can recover the
  current chat/sender without re-parsing the `tgbotapi.Update`.

Handlers have the signature `func(ctx context.Context, r *Request) (*Reply, error)` (`HandleFunc`). A
`*Reply` carries `[]tgbotapi.Chattable` to send; returning an `error` instead causes `handleUpdate` to
format and send a single error message using `go-kratos`'s `errors.FromError` instead of the reply.
`setMessageReplyDefaults`/`setChattableDefaults` in `bot.go` fill in chat ID, reply-to-message linking, and
default parse mode (`html`) on outgoing messages/callbacks before they're sent — handlers generally don't
need to set these themselves.

Logging/errors throughout the library use `github.com/go-kratos/kratos/v2/log` and
`github.com/go-kratos/kratos/v2/errors`, not the standard library — follow that convention when adding
new code (e.g. construct errors with `errors.New`/`errors.BadRequest` so they carry a reason code that
`formatError` can serialize).
