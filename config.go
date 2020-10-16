package telegrambot

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	Languages []string
}

func (c *Config) langFile(langName string) string {
	return fmt.Sprintf(`package languages

type %s struct {
}

func (%s %s) MainMenu() string {
	return "Welcome to Main Menu"
}`, strings.Title(langName), string(langName[0]), strings.Title(langName))
}

func (c *Config) langInterface() string {
	return `package languages

type Language interface {
	MainMenu() string
}
`
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
		f, err := os.OpenFile(langPath+"/"+language+".go", os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}
		content, _ := ioutil.ReadAll(f)
		str := string(content)
		lastBracket := strings.LastIndex(str, "}")
		f.Truncate(0)
		f.Seek(0, 0)
		f.WriteString(str[:lastBracket+1] + fmt.Sprintf(textContent, string(language[0]), strings.Title(language), text, strings.Join(arguments, ",")))
		f.Close()
	}
	return
}

func (c *Config) createLanguageFiles() (err error) {
	langPath := "languages"
	langInterfaceFilePath := langPath + "/interface.go"
	_, err = os.Stat(langInterfaceFilePath)
	if err == nil {
		return nil
	}
	if _, err := os.Stat(langPath); os.IsNotExist(err) {
		os.Mkdir(langPath, os.ModePerm)
	}
	o, err := os.OpenFile(langInterfaceFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	o.Write([]byte(c.langInterface()))
	o.Close()
	for _, language := range c.Languages {
		langPath := langPath + fmt.Sprintf("/%s.go", language)
		o, err := os.OpenFile(langPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
		if err != nil {
			return err
		}
		o.Write([]byte(c.langFile(language)))
		o.Close()
	}
	return nil
}
