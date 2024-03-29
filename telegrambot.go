package telegrambot

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	if config == nil {
		config = &Config{Languages: []string{"English"}}
	}
	gotelbbot = &Bot{application: application, Config: config}
	if gotelbbot.processFlags() {
		gotelbbot = nil
		client = nil
		err = errors.New("returned from flag process")
		return
	}
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
	if gotelbbot.Config != nil {
		gotelbbot.Config.createLanguageFiles()
	}
	return
}

func (gtb *Bot) ListenWebHook(address string) {
	go gtb.client.MethodByName("ListenWebhook").Call([]reflect.Value{
		reflect.ValueOf(address),
	})

	updatesChan := gtb.client.MethodByName("Updates").Call([]reflect.Value{})
	for update := range updatesChan[0].Interface().(chan *structs.Update) {
		fmt.Println("hereeee")
		go gtb.processUpdate(update)
	}
	return
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
		} else {
			app.MethodByName("ProcessGroupUpdate").Call([]reflect.Value{reflect.ValueOf(update)})
			return
		}
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

func (gtb *Bot) processFlags() (hasFlags bool) {
	var addText = flag.String("text", "", "--text=WelcomeMessage")
	var addLang = flag.String("lang", "", "--lang=Farsi")
	var init = flag.String("init", "", "--init=Bot")
	flag.Parse()

	if addLang != nil && *addLang != "" {
		hasFlags = true
		gtb.Config.addLanguage(*addLang)
		return
	}
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
			err = gtb.Config.addTextToLanguageFiles(*addText)
			if err == nil {
				exec.Command("gofmt", "-s", "-w", "./..").Run()
			}
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
