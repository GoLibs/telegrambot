package telegrambot

import (
	go_telegram_bot_api "github.com/GoLibs/telegram-bot-api"
	"github.com/GoLibs/telegram-bot-api/structs"
)

type Fields struct {
	Client     *go_telegram_bot_api.TelegramBot
	Update     *structs.Update
	IsSwitched bool
	SwitchMenu func(menu string) error
}

type Application interface {
	UserState() string
	OnUpdateHandlers(update *structs.Update)
	ProcessCallbackQuery(query *structs.CallbackQuery)
	MainMenu()
}
