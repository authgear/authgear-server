package template

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flosch/pongo2"
)

const MaxTemplateSize = 1024 * 1024 * 1

func DownloadTemplateFromFilePath(filePath string) (string, error) {
	filePath = filepath.Clean(filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(io.LimitReader(f, MaxTemplateSize))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func DownloadTemplateFromURL(url string) (string, error) {
	// FIXME(sec): validate URL to be trusted URL
	// nolint: gosec
	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return "", fmt.Errorf("unsuccessful request: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, MaxTemplateSize))
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ParseTextTemplateFromURL(url string, context map[string]interface{}) (string, error) {
	var body string
	var err error
	if body, err = DownloadTemplateFromURL(url); err != nil {
		return "", err
	}

	return ParseTextTemplate(body, context)
}

func ParseHTMLTemplateFromURL(url string, context map[string]interface{}) (string, error) {
	var body string
	var err error
	if body, err = DownloadTemplateFromURL(url); err != nil {
		return "", err
	}

	return ParseHTMLTemplate(body, context)
}

func ParseTextTemplate(templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	// turn off auto html escape
	autoEscapeOffTemplate := `{%% autoescape off %%}%s{%% endautoescape %%}`
	autoEscapeOffTemplateString := fmt.Sprintf(autoEscapeOffTemplate, templateString)

	return ParseHTMLTemplate(autoEscapeOffTemplateString, context)
}

func ParseHTMLTemplate(templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	var t *pongo2.Template
	if t, err = pongo2.FromString(templateString); err != nil {
		return
	}

	if out, err = t.Execute(context); err != nil {
		return
	}

	return
}
