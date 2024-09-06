package samlbinding

import (
	"encoding/base64"
	"html/template"
	"net/http"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPPostParseResult struct {
	AuthnRequestXML string
	AuthnRequest    *samlprotocol.AuthnRequest
	RelayState      string
}

var _ SAMLBindingParseResult = &SAMLBindingHTTPPostParseResult{}

func (*SAMLBindingHTTPPostParseResult) samlBindingParseResult() {}

func SAMLBindingHTTPPostParse(r *http.Request) (
	result *SAMLBindingHTTPPostParseResult,
	err error,
) {
	result = &SAMLBindingHTTPPostParseResult{}
	if err := r.ParseForm(); err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "failed to parse request body as a application/x-www-form-urlencoded form",
			Cause:  err,
		}
	}
	relayState := r.PostForm.Get("RelayState")
	result.RelayState = relayState

	requestBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
	if err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}

	result.AuthnRequestXML = string(requestBuffer)

	authnRequest, err := samlprotocol.ParseAuthnRequest(requestBuffer)
	if err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "malformed AuthnRequest",
			Cause:  err,
		}
	}
	result.AuthnRequest = authnRequest

	return result, nil
}

type SAMLBindingHTTPPostWriter struct{}

type postFormData struct {
	CallbackURL  string
	SAMLResponse string
	RelayState   string
}

const postForm = `
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

func (*SAMLBindingHTTPPostWriter) Write(
	rw http.ResponseWriter,
	callbackURL string,
	response *samlprotocol.Response,
	relayState string) error {

	responseEl := response.Element()

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)
	responseBuf, err := doc.WriteToBytes()
	if err != nil {
		return err
	}

	encodedResponse := base64.StdEncoding.EncodeToString(responseBuf)

	data := postFormData{
		CallbackURL:  callbackURL,
		SAMLResponse: encodedResponse,
		RelayState:   relayState,
	}

	tpl := template.Must(template.New("").Parse(postForm))
	if err := tpl.Execute(rw, data); err != nil {
		return err
	}
	return nil
}
