module github.com/unkmonster/tg-bot-server

go 1.25.5

require (
	github.com/go-kratos/kratos/v2 v2.9.2
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/samber/lo v1.53.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/OvyFlash/telegram-bot-api v0.0.0-20260715235732-aca8bf3898bb
