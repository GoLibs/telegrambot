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

func (a *App) ProcessGroupUpdate(update *structs.Update) {
	// commit
	return
}

func (a *App) ProcessCallbackQuery(query *structs.CallbackQuery) {
	return
}

func (a *App) RainMenu() {
	fmt.Println(a.IsSwitched)
	if !a.IsSwitched {
		fmt.Print("not switched")
		return
	}
	fmt.Println("switched")
	a.Client.Send(a.Client.Message().SetText("rain"))
}

func (a *App) MainMenu() {
	fmt.Println(a.IsSwitched)
	a.Client.Send(a.Client.Message().SetText("hi"))
	a.SwitchMenu("RainMenu")
}

func (a *App) UserState() string {
	return "MainMenu"
}

func (a *App) OnUpdateHandlers(update *structs.Update) {
}

func TestCreateBot(t *testing.T) {
	bot, _, err := telegrambot.NewBot("", &App{}, &telegrambot.Config{Languages: []string{"english", "farsi"}})
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
