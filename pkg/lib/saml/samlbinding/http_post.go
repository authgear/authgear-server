package samlbinding

import (
	"encoding/base64"
	"html/template"
	"net/http"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPPostParseRequestResult struct {
	SAMLRequestXML string
	RelayState     string
}

var _ SAMLBindingParseReqeustResult = &SAMLBindingHTTPPostParseRequestResult{}

func (*SAMLBindingHTTPPostParseRequestResult) samlBindingParseRequestResult() {}

func SAMLBindingHTTPPostParseRequest(r *http.Request) (
	result *SAMLBindingHTTPPostParseRequestResult,
	err error,
) {
	result = &SAMLBindingHTTPPostParseRequestResult{}
	if err := r.ParseForm(); err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "failed to parse request body as a application/x-www-form-urlencoded form",
			Cause:  err,
		}
	}
	relayState := r.PostForm.Get("RelayState")
	result.RelayState = relayState

	if r.PostForm.Get("SAMLRequest") == "" {
		return nil, ErrNoRequest
	}

	requestBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
	if err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}

	result.SAMLRequestXML = string(requestBuffer)

	return result, nil
}

type SAMLBindingHTTPPostParseResponseResult struct {
	SAMLResponseXML string
	RelayState      string
}

var _ SAMLBindingParseResponseResult = &SAMLBindingHTTPPostParseResponseResult{}

func (*SAMLBindingHTTPPostParseResponseResult) samlBindingParseResponseResult() {}

func SAMLBindingHTTPPostParseResponse(r *http.Request) (
	result *SAMLBindingHTTPPostParseResponseResult,
	err error,
) {
	result = &SAMLBindingHTTPPostParseResponseResult{}
	if err := r.ParseForm(); err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "failed to parse response body as a application/x-www-form-urlencoded form",
			Cause:  err,
		}
	}
	relayState := r.PostForm.Get("RelayState")
	result.RelayState = relayState

	if r.PostForm.Get("SAMLResponse") == "" {
		return nil, ErrNoResponse
	}

	responseBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLResponse"))
	if err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}

	result.SAMLResponseXML = string(responseBuffer)

	return result, nil
}

type SAMLBindingHTTPPostWriter struct{}

type responsePostFormData struct {
	CallbackURL  string
	SAMLResponse string
	RelayState   string
}

const responsePostForm = `
<html>
	<body onload="document.getElementById('f').submit();">
		<form method="POST" action="{{.CallbackURL}}" id="f">
			<input type="hidden" name="SAMLResponse" value="{{.SAMLResponse}}" />
			<input type="hidden" name="RelayState" value="{{.RelayState}}" />
			<noscript>
				<button type="submit">Continue</button>
			</noscript>
		</form>
	</body>
</html>
`

type requestPostFormData struct {
	CallbackURL string
	SAMLRequest string
	RelayState  string
}

const requestPostForm = `
<html>
	<body onload="document.getElementById('f').submit();">
		<form method="POST" action="{{.CallbackURL}}" id="f">
			<input type="hidden" name="SAMLRequest" value="{{.SAMLRequest}}" />
			<input type="hidden" name="RelayState" value="{{.RelayState}}" />
			<noscript>
				<button type="submit">Continue</button>
			</noscript>
		</form>
	</body>
</html>
`

func (*SAMLBindingHTTPPostWriter) WriteResponse(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	responseEl *etree.Element,
	relayState string) error {

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)
	responseBuf, err := doc.WriteToBytes()
	if err != nil {
		return err
	}

	encodedResponse := base64.StdEncoding.EncodeToString(responseBuf)

	data := responsePostFormData{
		CallbackURL:  callbackURL,
		SAMLResponse: encodedResponse,
		RelayState:   relayState,
	}

	tpl := template.Must(template.New("").Parse(responsePostForm))
	if err := tpl.Execute(rw, data); err != nil {
		return err
	}
	return nil
}

func (*SAMLBindingHTTPPostWriter) WriteRequest(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	requestEl *etree.Element,
	relayState string) error {

	doc := etree.NewDocument()
	doc.SetRoot(requestEl)
	requestBuf, err := doc.WriteToBytes()
	if err != nil {
		return err
	}

	encodedRequest := base64.StdEncoding.EncodeToString(requestBuf)

	data := requestPostFormData{
		CallbackURL: callbackURL,
		SAMLRequest: encodedRequest,
		RelayState:  relayState,
	}

	tpl := template.Must(template.New("").Parse(requestPostForm))
	if err := tpl.Execute(rw, data); err != nil {
		return err
	}
	return nil
}
