package tgbot

import (
	go_telegram_bot_api "github.com/GoLibs/telegram-bot-api"
	"github.com/GoLibs/telegram-bot-api/structs"
)

type Fields struct {
	Client *go_telegram_bot_api.TelegramBot
	Update *structs.Update
}

type Application interface {
	UserState() string
	OnUpdateHandlers(update *structs.Update)
	MainMenu()
}
