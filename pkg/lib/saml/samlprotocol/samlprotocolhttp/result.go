package samlprotocolhttp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SAMLResult struct {
	CallbackURL string
	Binding     samlprotocol.SAMLBinding
	Response    *samlprotocol.Response
	RelayState  string
}

func (s *SAMLResult) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	switch s.Binding {
	case samlprotocol.SAMLBindingHTTPPost:
		writer := &samlbinding.SAMLBindingHTTPPostWriter{}
		err := writer.Write(rw, s.CallbackURL, s.Response, s.RelayState)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("SAMLResult: unsupported binding %s", s.Binding))
	}
}

func (s *SAMLResult) IsInternalError() bool {
	// Not used
	return false
}

type SAMLErrorResult struct {
	SAMLResult
	Cause        error
	IsUnexpected bool
}

var _ error = &SAMLErrorResult{}

func (s *SAMLErrorResult) Error() string {
	return fmt.Sprintf("saml error response: %v", s.Cause)
}

func (s *SAMLErrorResult) Unwrap() error {
	return s.Cause
}

var _ httputil.Result = &SAMLErrorResult{}

func (s *SAMLErrorResult) IsInternalError() bool {
	return s.IsUnexpected
}

func NewExpectedSAMLErrorResult(cause error, result SAMLResult) *SAMLErrorResult {
	return &SAMLErrorResult{
		SAMLResult:   result,
		Cause:        cause,
		IsUnexpected: false,
	}
}

func NewUnexpectedSAMLErrorResult(cause error, result SAMLResult) *SAMLErrorResult {
	return &SAMLErrorResult{
		SAMLResult:   result,
		Cause:        cause,
		IsUnexpected: true,
	}
}
