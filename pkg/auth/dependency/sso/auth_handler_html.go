package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/template"
)

type AuthHandlerHTMLProvider struct {
	APIEndPoint *url.URL
}

func NewAuthHandlerHTMLProvider(APIEndPoint *url.URL) AuthHandlerHTMLProvider {
	return AuthHandlerHTMLProvider{
		APIEndPoint: APIEndPoint,
	}
}

func (i *AuthHandlerHTMLProvider) HTML(data map[string]interface{}) (out string, err error) {
	const templateString = `
<!DOCTYPE html>
<html>
<head>
<script type="text/javascript">

function postSSOResultMessageToWindow(windowObject) {
	var callbackURL = {{ .callback_url }};
	var result = {{ .result }};
	var error = null;
	if (!result) {
		error = 'Fail to retrieve result';
	} else if (!callbackURL) {
		error = 'Fail to retrieve callbackURL';
	}

	if (error) {
		windowObject.postMessage({
			type: "error",
			error: error
		}, "*");
	} else {
		windowObject.postMessage({
			type: "result",
			result: result
		}, callbackURL);
	}
	windowObject.postMessage({
		type: "end"
	}, "*");
}

if (window.opener) {
	postSSOResultMessageToWindow(window.opener);
} else {
	throw new Error("no window.opener");
}
</script>
</head>
<body>
</body>
</html>
	`
	context := map[string]interface{}{
		"api_endpoint": i.APIEndPoint.String(),
		"result":       data["result"],
		"callback_url": data["callback_url"],
	}

	return template.RenderHTMLTemplate(template.RenderOptions{
		Name:         "auth_handler",
		TemplateBody: templateString,
		Context:      context,
	})
}
