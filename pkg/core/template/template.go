package template

import (
	"fmt"

	"github.com/flosch/pongo2"
	"github.com/franela/goreq"
)

func DownloadTemplateFromURL(url string) (string, error) {
	req := goreq.Request{
		Method: "GET",
		Uri:    url,
	}

	var err error
	var resp *goreq.Response
	if resp, err = req.Do(); err != nil {
		return "", err
	}

	var body string
	if body, err = resp.Body.ToString(); err != nil {
		return "", err
	}

	return body, nil
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
