package welcemail

import (
	"fmt"

	"github.com/flosch/pongo2"
	"github.com/franela/goreq"
)

func downloadTemplateFromURL(url string) (string, error) {
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

func parseTextTemplateFromURL(url string, context map[string]interface{}) (string, error) {
	var body string
	var err error
	if body, err = downloadTemplateFromURL(url); err != nil {
		return "", err
	}

	return parseTextTemplate(body, context)
}

func parseHTMLTemplateFromURL(url string, context map[string]interface{}) (string, error) {
	var body string
	var err error
	if body, err = downloadTemplateFromURL(url); err != nil {
		return "", err
	}

	return parseHTMLTemplate(body, context)
}

func parseTextTemplate(templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	// turn off auto html escape
	autoEscapeOffTemplate := `{%% autoescape off %%}%s{%% endautoescape %%}`
	autoEscapeOffTemplateString := fmt.Sprintf(autoEscapeOffTemplate, templateString)

	return parseHTMLTemplate(autoEscapeOffTemplateString, context)
}

func parseHTMLTemplate(templateString string, context map[string]interface{}) (out string, err error) {
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
