package samlprotocolhttp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type ResultWriterSigner interface {
	ConstructSignedQueryParameters(
		samlResponse string,
		relayState string,
	) (url.Values, error)
}

type ResultWriter struct {
	Signer ResultWriterSigner

	HTTPPostWriter     *samlbinding.SAMLBindingHTTPPostWriter
	HTTPRedirectWriter *samlbinding.SAMLBindingHTTPRedirectWriter
}

type WriteOptions struct {
	Binding     samlprotocol.SAMLBinding
	CallbackURL string
	RelayState  string
}

func (w *ResultWriter) Write(
	rw http.ResponseWriter,
	r *http.Request,
	result SAMLResult, options *WriteOptions) error {

	switch options.Binding {
	case samlprotocol.SAMLBindingHTTPPost:
		err := w.HTTPPostWriter.Write(rw, r, options.CallbackURL, result.GetResponse(), options.RelayState)
		if err != nil {
			return err
		}
	case samlprotocol.SAMLBindingHTTPRedirect:
		err := w.HTTPRedirectWriter.Write(rw, r, options.CallbackURL, result.GetResponse(), options.RelayState)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("ResultWriter: unsupported binding %s", options.Binding)
	}
	return nil
}
