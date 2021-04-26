package telegrambot

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Config struct {
	Languages []string
}

type langFns struct {
	Name    string
	Inputs  string
	Outputs string
}

func (c *Config) addLangData(langName string, fns []langFns) string {
	var fnStrings []string
	for _, fn := range fns {
		s := fmt.Sprintf(`
	func (%s %s) %s(%s) %s {
		return ""
	}
`, strings.ToLower(langName[0:1]), strings.Title(langName), fn.Name, fn.Inputs, fn.Outputs)
		fnStrings = append(fnStrings, s)
	}
	return fmt.Sprintf(`package languages

type %s struct {

}

%s
`, strings.Title(langName), strings.Join(fnStrings, "\r\n"))
}

func (c *Config) langFileData(langName string) string {
	return fmt.Sprintf(`package languages

type %s struct {
}

func (%s %s) MainMenu() string {
	return "Welcome to Main Menu"
}`, strings.Title(langName), strings.ToLower(string(langName[0])), strings.Title(langName))
}

func (c *Config) langInterfaceData() string {
	return `package languages

type Language interface {
	MainMenu() string
}
`
}

func (c *Config) appData(appName string) string {
	appName = strings.ToUpper(appName[0:1]) + strings.ToLower(appName[1:])
	shortName := strings.ToLower(appName[0:1])
	text := `package app

import (
	"github.com/GoLibs/telegram-bot-api/structs"
	"github.com/GoLibs/telegrambot"
)

type %NAME% struct {
	telegrambot.Fields
	l    languages.Language
	User interface{} // TODO: Set User Model Here 
}

func (%SHORT% *%NAME%) UserState() string {
	return "MainMenu"
}

func (%SHORT% *%NAME%) OnUpdateHandlers(update *structs.Update) {
	%SHORT%.l = languages.English{}
	if update.Message == nil {
		return
	}
	if update.Message.From != nil && %SHORT%.User == nil {
		// TODO: Create or Initialise User Here
	} else {
		// TODO: Refresh User State Here
	}
}

func  (%SHORT% *%NAME%) ProcessCallbackQuery(query *structs.CallbackQuery) {
	// commit
}

func  (%SHORT% *%NAME%) ProcessGroupUpdate(update *structs.Update) {
	// commit
}

func(%SHORT% *%NAME%) MainMenu() {
	// TODO: Set User State
	if !%SHORT%.IsSwitched {
		// TODO:
	}
	%SHORT%.Client.Send(%SHORT%.Client.Message().SetText("Hi"))
}
`
	text = strings.ReplaceAll(text, "%SHORT%", shortName)
	text = strings.ReplaceAll(text, "%NAME%", appName)
	return text
}

func (c *Config) addTextToLanguageFiles(text string) (err error) {
	var arguments []string
	split := strings.Split(text, ",")
	if len(split) > 1 {
		text = split[0]
		arguments = split[1:]
	}
	langPath := "languages"
	textContent := `

func (%s %s) %s(%s) string {
	return ""
}
`

	for _, language := range c.Languages {
		f, err := os.OpenFile(langPath+string(os.PathSeparator)+language+".go", os.O_RDWR, 644)
		if err != nil {
			return err
		}
		content, _ := ioutil.ReadAll(f)
		str := string(content)
		lastBracket := strings.LastIndex(str, "}")
		f.Truncate(0)
		f.Seek(0, 0)
		f.WriteString(str[:lastBracket+1] + fmt.Sprintf(textContent, strings.ToLower(string(language[0])), strings.Title(language), text, strings.Join(arguments, ",")))
		f.Close()
	}
	return
}

func (c *Config) createLanguageFiles() (err error) {
	langPath := "languages"
	langInterfaceFilePath := langPath + string(os.PathSeparator) + "interface.go"
	_, err = os.Stat(langInterfaceFilePath)
	if err == nil {
		return nil
	}
	if _, err := os.Stat(langPath); os.IsNotExist(err) {
		os.Mkdir(langPath, 644)
	}
	o, err := os.OpenFile(langInterfaceFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 644)
	if err != nil {
		return err
	}
	o.Write([]byte(c.langInterfaceData()))
	o.Close()
	for _, language := range c.Languages {
		langPath := langPath + string(os.PathSeparator) + fmt.Sprintf("%s.go", strings.ToLower(language))
		o, err := os.OpenFile(langPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 644)
		if err != nil {
			return err
		}
		o.Write([]byte(c.langFileData(language)))
		o.Close()
	}
	return nil
}

func (c *Config) init(appName string) (err error) {
	c.createLanguageFiles()
	appPath := "app"
	if _, err = os.Stat(appPath); os.IsNotExist(err) {
		err = os.MkdirAll(appPath, 644)
		if err != nil {
			return
		}
	}
	filename := strings.ToLower(appName)
	err = ioutil.WriteFile("app/"+filename+".go", []byte(c.appData(appName)), 644)
	if err == nil {
		exec.Command("gofmt", "-s", "-w", "./..").Run()
	}
	return
}

func (c *Config) addLanguage(langName string) (err error) {
	langPath := "languages"
	interfacePath := langPath + string(os.PathSeparator) + "interface.go"
	addLangPath := langPath + string(os.PathSeparator) + strings.ToLower(langName) + ".go"
	var data []byte
	data, err = ioutil.ReadFile(interfacePath)
	if err != nil {
		return
	}
	var fns []langFns
	r := regexp.MustCompile(`[\s\S.]+?(.+?)\((.*?)\)\s+?(.*)`)
	for _, fn := range r.FindAllStringSubmatch(string(data), -1) {
		fnName := fn[1]
		args := fn[2]
		output := fn[3]
		fns = append(fns, langFns{
			Name:    fnName,
			Inputs:  args,
			Outputs: output,
		})
	}
	err = ioutil.WriteFile(addLangPath, []byte(c.addLangData(langName, fns)), os.ModePerm)
	if err == nil {
		exec.Command("gofmt", "-s", "-w", "./..").Run()
	}
	return
}
