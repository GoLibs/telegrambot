package tests

import (
	"fmt"
	"testing"

	"github.com/GoLibs/telegrambot"

	"github.com/GoLibs/telegram-bot-api/structs"
)

type App struct {
	telegrambot.Fields
}

func (a *App) RainMenu() {
	if !a.IsSwitched {
		fmt.Print("not switched")
		return
	}
	fmt.Print("switched")
	a.Client.Send(a.Client.Message().SetText("rain"))
}

func (a *App) MainMenu() {
	fmt.Print("switched")
	a.Client.Send(a.Client.Message().SetText("hi"))
	a.SwitchMenu("RainMenu")
}

func (a *App) UserState() string {

	return "MainMenu"
}

func (a *App) OnUpdateHandlers(update *structs.Update) {
}

func TestCreateBot(t *testing.T) {
	bot, err := telegrambot.NewBot("", &App{})
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
