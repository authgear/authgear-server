package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/authgear/authgear-server/pkg/util/intl"
)

type Locale struct {
	Name            string
	TranslationFile string
}

var MESSAGES_PATH string = "messages"
var TRANSLATION_FILE_PATH string = "translation.json"

func CutSuffix(s, suffix string) (before string, found bool) {
	if !strings.HasSuffix(s, suffix) {
		return s, false
	}
	return s[:len(s)-len(suffix)], true
}

func cloneMap(original map[string]any) map[string]any {
	copied := make(map[string]any)
	for k, v := range original {
		copied[k] = v
	}
	return copied
}

func constructTemplatePath(templatesDirectory string, fileName string) string {
	return filepath.Join(templatesDirectory, fileName)
}

func getLocales(
	templatesDirectory string) []Locale {
	locales := []Locale{}

	dirs, err := os.ReadDir(templatesDirectory)
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			locale := dir.Name()
			locales = append(locales, Locale{
				Name:            locale,
				TranslationFile: filepath.Join(templatesDirectory, locale, MESSAGES_PATH, TRANSLATION_FILE_PATH),
			})
		}
	}

	return locales
}

func renderLocalizedTemplate(
	template string,
	locale Locale,
	templatesDirectory string,
	defaultTemplatesDirectory string) {
	translationJson, err := loadJson(locale.TranslationFile)
	if err != nil {
		panic(err)
	}
	context := make(map[string]any)
	// mjml > 4.14.1 <= 4.15.2 has a bug that always outputs lang="und" and dir="auto"
	// So we tell mjml the lang and the dir.
	// See https://github.com/mjmlio/mjml/issues/2865
	context["lang"] = locale.Name
	context["dir"] = intl.HTMLDir(locale.Name)
	context["T"] = translationJson
	renderTemplate(template, templatesDirectory, defaultTemplatesDirectory, context)
}

func renderTemplate(
	template string,
	templatesDirectory string,
	defaultTemplatesDirectory string,
	context any) {
	var fullBuffer strings.Builder
	var plaintextBuffer strings.Builder

	tplName, found := CutSuffix(filepath.Base(template), ".gotemplate")
	if !found {
		panic("render: invalid template path " + template)
	}

	baseTemplate := getBaseTemplate(&plaintextBuffer)
	tpl, err := loadTemplate(constructTemplatePath(templatesDirectory, template), baseTemplate)
	if errors.Is(err, fs.ErrNotExist) {
		tpl, err = loadTemplate(constructTemplatePath(defaultTemplatesDirectory, template), baseTemplate)
	}
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(&fullBuffer, filepath.Base(template), context)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(constructTemplatePath(templatesDirectory, tplName), []byte(fullBuffer.String()), 0666)
	if err != nil {
		panic(err)
	}
}

func getBaseTemplate(plaintextBuffer *strings.Builder) *template.Template {
	funcMap := template.FuncMap{
		"plaintext": func(ss ...string) string {
			str := strings.Join(ss, " ")
			if str != "" {
				plaintextBuffer.WriteString(str + "\n")
			}
			return str
		},
	}

	return template.New("").Delims("[[", "]]").Funcs(funcMap)
}

func loadTemplate(path string, baseTemplate *template.Template) (*template.Template, error) {
	return baseTemplate.ParseFiles(path)
}

func main() {
	templatesDirectory := flag.String("i", "templates", "template files directory")
	flag.Parse()

	defaultMessagesDir := filepath.Join(*templatesDirectory, "en", MESSAGES_PATH)

	var messageTemplates []string
	err := filepath.WalkDir(defaultMessagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".gotemplate" {
			template, err := filepath.Rel(defaultMessagesDir, path)
			if err != nil {
				panic(err)
			}
			messageTemplates = append(messageTemplates, template)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	dirs, err := os.ReadDir(*templatesDirectory)
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		locale := dir.Name()
		localeMessagesDir := filepath.Join(*templatesDirectory, locale, MESSAGES_PATH)
		if _, err := os.Stat(localeMessagesDir); os.IsNotExist(err) {
			continue
		}

		localeTranslationFile := filepath.Join(localeMessagesDir, TRANSLATION_FILE_PATH)
		if _, err := os.Stat(localeTranslationFile); os.IsNotExist(err) {
			continue
		}

		for _, template := range messageTemplates {
			renderLocalizedTemplate(template, Locale{
				Name:            locale,
				TranslationFile: localeTranslationFile,
			}, localeMessagesDir, defaultMessagesDir)
		}
	}
}

func loadJson(path string) (map[string]any, error) {
	rawConfig, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
