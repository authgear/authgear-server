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

function validateCallbackURL(callbackURL, authorizedURLs) {
	if (!callbackURL) {
		return false;
	}
	for (var i = 0; i < authorizedURLs.length; ++i) {
		if (callbackURL === authorizedURLs[i]) {
			return true;
		}
	}
	return false;
}

function postSSOResultMessageToWindow(windowObject, authorizedURLs) {
	var callbackURL = {{ .callback_url }};
	var result = {{ .result }};
	var error = null;
	if (!result) {
		error = 'Fail to retrieve result';
	} else if (!callbackURL) {
		error = 'Fail to retrieve callbackURL';
	} else if (!validateCallbackURL(callbackURL, authorizedURLs)) {
		error = "Unauthorized callback URL: " + callbackURL;
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

var req = new XMLHttpRequest();
req.onload = function() {
	var jsonResponse = JSON.parse(req.responseText);
	var authorizedURLs = jsonResponse.result.authorized_urls;
	if (window.opener) {
		postSSOResultMessageToWindow(window.opener, authorizedURLs);
	} else {
		throw new Error("no window.opener");
	}
};
req.open("POST", "{{ .api_endpoint }}/_auth/sso/config", true);
req.send(null);
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

	return template.RenderHTMLTemplate("auth_handler", templateString, context)
}
