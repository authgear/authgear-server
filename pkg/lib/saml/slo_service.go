package saml

import (
	"net/http"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlslosession"
)

type BindingHTTPPostWriter interface {
	WriteRequest(
		rw http.ResponseWriter,
		r *http.Request,
		callbackURL string,
		requestElement *etree.Element,
		relayState string) error
}

type BindingHTTPRedirectWriter interface {
	WriteRequest(
		rw http.ResponseWriter,
		r *http.Request,
		callbackURL string,
		requestElement *etree.Element,
		relayState string) error
}

type SAMLService interface {
	IssueLogoutRequest(
		sp *config.SAMLServiceProviderConfig,
		sloSession *samlslosession.SAMLSLOSession,
	) (*samlprotocol.LogoutRequest, error)
}

type SLOService struct {
	SAMLService               SAMLService
	BindingHTTPPostWriter     BindingHTTPPostWriter
	BindingHTTPRedirectWriter BindingHTTPRedirectWriter
}

func (s *SLOService) SendSLORequest(
	rw http.ResponseWriter,
	r *http.Request,
	sloSession *samlslosession.SAMLSLOSession,
	sp *config.SAMLServiceProviderConfig,
) error {
	logoutRequest, err := s.SAMLService.IssueLogoutRequest(
		sp,
		sloSession,
	)
	if err != nil {
		return err
	}
	logoutRequestEl := logoutRequest.Element()
	callbackURL := sp.SLOCallbackURL
	switch sp.SLOBinding {
	case samlprotocol.SAMLBindingHTTPPost:
		err = s.BindingHTTPPostWriter.WriteRequest(rw, r, callbackURL, logoutRequestEl, sloSession.ID)
		if err != nil {
			return err
		}
	case samlprotocol.SAMLBindingHTTPRedirect:
		err = s.BindingHTTPRedirectWriter.WriteRequest(rw, r, callbackURL, logoutRequestEl, sloSession.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
