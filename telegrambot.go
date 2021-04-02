package telegrambot

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	go_telegram_bot_api "github.com/GoLibs/telegram-bot-api"

	"github.com/GoLibs/telegram-bot-api/structs"
)

type Bot struct {
	application      Application
	applicationValue reflect.Value
	applicationType  reflect.Type
	client           reflect.Value
	Config           *Config
}

func NewBot(token string, application Application, config *Config) (gotelbbot *Bot, client *go_telegram_bot_api.TelegramBot, err error) {
	appVal := reflect.ValueOf(application)

	fields := appVal.Elem().FieldByName("Fields")
	if !fields.IsValid() {
		err = errors.New("fields_not_found")
		return
	}
	if _, ok := fields.Interface().(Fields); !ok {
		err = errors.New("fields_not_found")
		return
	}

	clientField := appVal.Elem().FieldByName("Client")
	if !clientField.IsValid() {
		err = errors.New("client_field_not_found")
		return
	}

	updateField := appVal.Elem().FieldByName("Update")
	if !updateField.IsValid() {
		err = errors.New("update_field_not_found")
		return
	}
	gotelbbot = &Bot{application: application}
	if gotelbbot.ProcessFlags() {
		gotelbbot = nil
		client = nil
		err = nil
		return
	}
	client, err = go_telegram_bot_api.NewTelegramBot(token)
	if err != nil {
		return
	}
	clientField.Set(reflect.ValueOf(client))
	gotelbbot.applicationType = reflect.TypeOf(application)
	gotelbbot.applicationValue = appVal

	switchMenuField := appVal.Elem().FieldByName("SwitchMenu")
	switchMenuField.Set(reflect.ValueOf(gotelbbot.SwitchMenu))

	gotelbbot.client = clientField
	gotelbbot.Config = config
	if gotelbbot.Config != nil {
		gotelbbot.Config.createLanguageFiles()
	}
	return
}

func (gtb *Bot) ListenWebHook(address string) {
	gtb.client.MethodByName("GetUpdates").Call([]reflect.Value{})
}

func (gtb *Bot) GetUpdates() error {
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

func (gtb *Bot) SwitchMenu(menuName string) error {
	menu := gtb.applicationValue.MethodByName(menuName)
	if !menu.IsValid() {
		return errors.New("menu_not_found")
	}
	gtb.applicationValue.Elem().FieldByName("Fields").FieldByName("IsSwitched").Set(reflect.ValueOf(true))
	menu.Call([]reflect.Value{})
	return nil
}

func (gtb *Bot) processUpdate(update *structs.Update) {
	app := reflect.New(gtb.applicationType.Elem())
	appValue := app.Elem()
	client := appValue.FieldByName("Client")
	client.Set(gtb.client)
	updateField := appValue.FieldByName("Update")
	updateField.Set(reflect.ValueOf(update))
	switchMenu := func(menuName string) error {
		menu := app.MethodByName(menuName)
		if !menu.IsValid() {
			return errors.New("menu_not_found")
		}
		appValue.FieldByName("IsSwitched").Set(reflect.ValueOf(true))
		menu.Call([]reflect.Value{})
		return nil
	}

	switchMenuField := appValue.FieldByName("SwitchMenu")
	switchMenuField.Set(reflect.ValueOf(switchMenu))

	app.MethodByName("OnUpdateHandlers").Call([]reflect.Value{reflect.ValueOf(update)})
	var chat *structs.Chat
	if update.Message != nil {
		/*if Update.Message.From != nil {
			gtb.applicationValue.MethodByName("SetMessageUser").Call([]reflect.Value{reflect.ValueOf(Update.Message.From)})
		}*/
		chat = update.Message.Chat

		appValue.FieldByName("Client").MethodByName("SetRecipientChatId").Call([]reflect.Value{reflect.ValueOf(chat.Id)})
		if chat.Type == "private" {
			gtb.processMenu(app)
			return
		} /* else {
			application.ProcessGroupUpdate()
			return
		}*/
	} else if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		chat = update.CallbackQuery.Message.Chat
		app.MethodByName("ProcessCallbackQuery").Call([]reflect.Value{reflect.ValueOf(update.CallbackQuery)})
		return
	}
}

func (gtb *Bot) processMenu(applicationValue reflect.Value) {
	// applicationValue := reflect.ValueOf(application)
	menu := applicationValue.MethodByName("UserState").Call([]reflect.Value{})[0].Interface().(string)

	_, ok := gtb.applicationType.MethodByName(menu)
	if ok {
		applicationValue.MethodByName(menu).Call([]reflect.Value{})
		return
	}
	applicationValue.MethodByName("MainMenu").Call([]reflect.Value{})
}

func (gtb *Bot) ProcessFlags() (hasFlags bool) {
	var addText = flag.String("text", "", "--text=WelcomeMessage")
	var init = flag.String("init", "", "--init=Bot")
	flag.Parse()

	if addText != nil && *addText != "" {
		hasFlags = true
		langPath := "languages"
		interfacePath := langPath + "/interface.go"
		langInterfaceFile, err := os.OpenFile(interfacePath, os.O_RDWR, 644)
		if err != nil {
			return
		}
		langInterface, err := ioutil.ReadAll(langInterfaceFile)
		if err != nil {
			return
		}
		str := string(langInterface)
		lastBracket := strings.LastIndex(str, "}") - 1

		var arguments []string
		split := strings.Split(*addText, ",")
		textContent := split[0]
		if len(split) > 1 {
			arguments = split[1:]
		}
		str = str[:lastBracket] + "\n" + textContent + fmt.Sprintf("(%s) string", strings.Join(arguments, ",")) + "\n}"
		langInterfaceFile.Truncate(0)
		langInterfaceFile.Seek(0, 0)
		langInterfaceFile.WriteString(str)
		langInterfaceFile.Close()
		if gtb.Config != nil {
			gtb.Config.addTextToLanguageFiles(*addText)
		}
		return
	}
	if init != nil && *init != "" {
		hasFlags = true
		if gtb.Config != nil {
			err := gtb.Config.init(*init)
			if err != nil {
				fmt.Println(err)
			}
		}
		return
	}
	return
}
