package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Locale struct {
	Name            string
	TranslationFile string
}

var MESSAGES_PATH string = "messages"

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

func constructOutputPath(outputDirectory string, locale Locale, path string) string {
	return filepath.Join(outputDirectory, "templates", locale.Name, path)
}

func constructInputPath(templatesDirectory string, templatePath string, fileName string) string {
	return filepath.Join(templatesDirectory, templatePath, fileName)
}

func getLocales(
	tranlationsDirectory string) []Locale {
	locales := []Locale{}
	err := filepath.Walk(tranlationsDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fileName := info.Name()
		if !strings.HasSuffix(fileName, ".json") {
			return nil
		}
		locale := strings.TrimSuffix(fileName, ".json")
		locales = append(locales, Locale{
			Name:            locale,
			TranslationFile: path,
		})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return locales
}

func renderLocalizedTemplate(
	template string,
	locales []Locale,
	templatesDirectory string,
	outputDirectory string,
	templatePath string) {

	for _, locale := range locales {
		localeOutputDir := constructOutputPath(outputDirectory, locale, templatePath)
		translationJson, err := loadJson(locale.TranslationFile)
		if err != nil {
			panic(err)
		}
		context := make(map[string]any)
		context["T"] = translationJson
		renderTemplate(template, templatesDirectory, localeOutputDir, templatePath, context)
	}
}

func renderTemplate(
	template string,
	templatesDirectory string,
	outputDirectory string,
	templatePath string,
	context any) {
	var fullBuffer strings.Builder
	var plaintextBuffer strings.Builder

	tplName, found := CutSuffix(filepath.Base(template), ".gotemplate")
	if !found {
		panic("render: invalid template path " + template)
	}

	baseTemplate := getBaseTemplate(&plaintextBuffer)
	tpl, err := loadTemplate(constructInputPath(templatesDirectory, templatePath, template), baseTemplate)
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(&fullBuffer, filepath.Base(template), context)
	if err != nil {
		panic(err)
	}

	outDir := filepath.Join(outputDirectory, filepath.Dir(template))
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(outDir, tplName), []byte(fullBuffer.String()), 0666)
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
		"url": func(ss ...string) string {
			str := strings.Join(ss, "")
			if str != "" {
				plaintextBuffer.WriteString(str + "\n")
			}
			return str
		},
		"formattedplaintext": func(s string, data ...string) string {
			str := s
			if data != nil {
				for idx, value := range data {
					str = strings.ReplaceAll(str, fmt.Sprintf("{%d}", idx), value)
				}
			}
			if str != "" {
				plaintextBuffer.WriteString(str + "\n")
			}
			return str
		},
		"concat": func(ss ...string) string {
			return strings.Join(ss, "")
		},
	}

	return template.New("").Delims("[[", "]]").Funcs(funcMap)
}

func loadTemplate(path string, baseTemplate *template.Template) (*template.Template, error) {
	return baseTemplate.ParseFiles(path)
}

func main() {
	tranlationsDirectory := flag.String("t", "translations", "translation files directory")
	templatesDirectory := flag.String("i", "templates", "template files directory")
	outputDirectory := flag.String("o", "output", "output directory path")
	flag.Parse()

	messageInputDir := filepath.Join(*templatesDirectory, MESSAGES_PATH)

	var messageTemplates []string
	err := filepath.WalkDir(messageInputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".gotemplate" {
			template, err := filepath.Rel(messageInputDir, path)
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

	locales := getLocales(*tranlationsDirectory)

	for _, tplFile := range messageTemplates {
		renderLocalizedTemplate(tplFile, locales, *templatesDirectory, *outputDirectory, MESSAGES_PATH)
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
