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

func (i *AuthHandlerHTMLProvider) HTML() (out string, err error) {
	const templateString = `
<!DOCTYPE html>
<html>
<head>
<script type="text/javascript">
function cookieJarGet(name) {
	var jar = [];
	var pairs = document.cookie.split("; ");
	var i = 0;
	for (; i < pairs.length; ++i) {
		var parts = pairs[i].split("=");
		var key = parts[0];
		var value = parts.slice(1).join("=");
		jar.push([key, value]);
	}
	for (var i = 0; i < jar.length; ++i) {
		if (jar[i][0] === name) {
			return jar[i][1];
		}
	}
}

function cookieJarDelete(name) {
	document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:01 GMT;";
}

function StringStartsWith(s, search) {
	return s.substring(0, search.length) === search;
}

function validateCallbackURL(callbackURL, authorizedURLs) {
	if (!callbackURL) {
		return false;
	}
	for (var i = 0; i < authorizedURLs.length; ++i) {
		if (StringStartsWith(callbackURL, authorizedURLs[i])) {
			return true;
		}
	}
	return false;
}

function postSSOResultMessageToWindow(windowObject, authorizedURLs) {
	var resultStr = cookieJarGet("sso_data");
	cookieJarDelete("sso_data");
	var data = resultStr && JSON.parse(atob(resultStr));
	var callbackURL = data && data.callback_url;
	var result = data && data.result;
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
	}

	return template.RenderHTMLTemplate("auth_handler", templateString, context)
}
