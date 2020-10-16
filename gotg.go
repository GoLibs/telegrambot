package tgbot

import (
	"errors"
	"reflect"

	go_telegram_bot_api "github.com/GoLibs/telegram-bot-api"

	"github.com/GoLibs/telegram-bot-api/structs"
)

type GoTelBot struct {
	application      Application
	applicationValue reflect.Value
	applicationType  reflect.Type
	client           reflect.Value
	update           reflect.Value
}

func NewGoTelBot(token string, application Application) (gotelbbot *GoTelBot, err error) {
	appVal := reflect.ValueOf(application)
	el := appVal.Elem().FieldByName("Client")
	if !el.IsValid() {
		err = errors.New("client_field_not_found")
		return
	}
	updateField := appVal.Elem().FieldByName("Update")
	if !updateField.IsValid() {
		err = errors.New("update_field_not_found")
		return
	}

	var client *go_telegram_bot_api.TelegramBot
	client, err = go_telegram_bot_api.NewTelegramBot(token)
	if err != nil {
		return
	}
	el.Set(reflect.ValueOf(client))
	gotelbbot = &GoTelBot{application: application}
	gotelbbot.applicationType = reflect.TypeOf(application)
	gotelbbot.applicationValue = appVal
	gotelbbot.client = el
	gotelbbot.update = updateField
	return
}

func (gtb *GoTelBot) ListenWebHook(address string) {
	gtb.client.MethodByName("GetUpdates").Call([]reflect.Value{})
}

func (gtb *GoTelBot) GetUpdates() error {
	getUpdates := gtb.client.MethodByName("GetUpdates").Call([]reflect.Value{})
	values := gtb.client.MethodByName("GetUpdatesChannel").Call([]reflect.Value{getUpdates[0]})
	updates := values[0].Interface().(structs.UpdatesChannel)
	err := values[1].Interface()
	if err != nil {
		return err.(error)
	}
	for update := range updates {
		go gtb.processUpdate(&update)
	}
	return nil
}

func (gtb *GoTelBot) processUpdate(update *structs.Update) {
	application := gtb.application
	gtb.update.Set(reflect.ValueOf(update))
	var chat *structs.Chat
	if update.Message != nil {
		/*if Update.Message.From != nil {
			gtb.applicationValue.MethodByName("SetMessageUser").Call([]reflect.Value{reflect.ValueOf(Update.Message.From)})
		}*/
		chat = update.Message.Chat
		gtb.client.MethodByName("SetRecipientChatId").Call([]reflect.Value{reflect.ValueOf(chat.Id)})
		if chat.Type == "private" {
			application.OnUpdateHandlers(update)
			gtb.processMenu(application)
			return
		} /* else {
			application.ProcessGroupUpdate()
			return
		}*/
	} else if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		chat = update.Message.Chat
	}
}

func (gtb *GoTelBot) processMenu(application Application) {
	applicationValue := reflect.ValueOf(application)
	menu := application.UserState()
	_, ok := gtb.applicationType.MethodByName(menu)
	if ok {
		applicationValue.MethodByName(menu).Call([]reflect.Value{})
		return
	}
	applicationValue.MethodByName("MainMenu").Call([]reflect.Value{})
}
