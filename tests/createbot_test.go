package tests

import (
	"fmt"
	"testing"
	"tgbot"

	"github.com/kr/pretty"

	"github.com/GoLibs/telegram-bot-api/structs"
)

type App struct {
	tgbot.Fields
}

func (a App) MainMenu() {
	a.Client.Send(a.Client.Message().SetText("hi"))
}

func (a App) UserState() string {
	return "MainMenu"
}

func (a App) OnUpdateHandlers(update *structs.Update) {
	pretty.Println(update)
}

func TestCreateBot(t *testing.T) {
	bot, err := tgbot.NewGoTelBot("", &App{})
	if err != nil {
		fmt.Print(err)
		return
	}
	err = bot.GetUpdates()
	if err != nil {
		fmt.Print(err)
	}
	return
}
