package saml

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlslosession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/beevik/etree"
)

func ConfigureLogoutRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/saml2/logout/:service_provider_id")
}

type LogoutHandlerLogger struct{ *log.Logger }

func NewLogoutHandlerLogger(lf *log.Factory) *LogoutHandlerLogger {
	return &LogoutHandlerLogger{lf.New("saml-logout-handler")}
}

type logoutResult interface {
	logoutResult()
}

// logoutCompleteResult means the whole logout proccess is completed,
// and a response should be returned
type logoutCompleteResult struct {
	response *etree.Element
}

func (*logoutCompleteResult) logoutResult() {}

var _ logoutResult = &logoutCompleteResult{}

// logoutRemainingSPsResult means we have to logout remaining SPs before completing the logout
type logoutRemainingSPsResult struct {
	sloSession *samlslosession.SAMLSLOSession
}

func (*logoutRemainingSPsResult) logoutResult() {}

type LogoutHandler struct {
	Logger                *LogoutHandlerLogger
	Clock                 clock.Clock
	SAMLConfig            *config.SAMLConfig
	SAMLService           HandlerSAMLService
	SessionManager        SessionManager
	SAMLSLOSessionService SAMLSLOSessionService

	BindingHTTPPostWriter     BindingHTTPPostWriter
	BindingHTTPRedirectWriter BindingHTTPRedirectWriter
}

func (h *LogoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	sp, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	if !sp.SLOEnabled {
		http.NotFound(rw, r)
		return
	}

	callbackURL := sp.SLOCallbackURL
	responseBinding := sp.SLOBinding
	var relayState string

	defer func() {
		if err := recover(); err != nil {
			e := panicutil.MakeError(err)
			h.handleError(rw, r, responseBinding, callbackURL, relayState, e)
		}
	}()

	var result logoutResult
	var err error
	relayState, result, err = h.handleSLORequest(rw, r, sp, responseBinding, callbackURL)
	if err != nil {
		h.handleError(rw, r, responseBinding, callbackURL, relayState, err)
		return
	}

	switch result := result.(type) {
	case *logoutCompleteResult:
		// Finish the logout with a response
		h.writeResponse(rw, r, result.response, responseBinding, callbackURL, relayState)
		return
	case *logoutRemainingSPsResult:
		// Logout all remaining participants before finish
		isSomeFailed := false
		for _, spID := range result.sloSession.Entry.PendingLogoutServiceProviderIDs.Keys() {
			sp, ok := h.SAMLConfig.ResolveProvider(spID)
			if ok && sp.SLOEnabled {
				err = h.sendSLORequest(
					rw, r,
					result.sloSession,
					sp,
				)
				if err != nil {
					// For some reason it failed
					// Skip this SP and send request to the next one
					isSomeFailed = true
					continue
				}
				return
			}
		}
		// None of the SPs has slo enabled, end the logout immediately
		logoutResponse, err := h.SAMLService.IssueLogoutResponse(
			callbackURL,
			sp.GetID(),
			result.sloSession.Entry.LogoutRequest(),
			isSomeFailed,
		)
		if err != nil {
			h.handleError(rw, r,
				result.sloSession.Entry.ResponseBinding,
				result.sloSession.Entry.CallbackURL,
				result.sloSession.Entry.RelayState,
				err,
			)
			return
		}
		h.writeResponse(rw, r,
			logoutResponse.Element(),
			result.sloSession.Entry.ResponseBinding,
			result.sloSession.Entry.CallbackURL,
			result.sloSession.Entry.RelayState,
		)
		return

	}
}

func (h *LogoutHandler) parseSLORequest(
	r *http.Request,
) (
	parseResult samlbinding.SAMLBindingParseResult,
	logoutRequest *samlprotocol.LogoutRequest,
	relayState string,
	err error,
) {
	defer func() {
		// Transform known errors
		if err != nil {
			var parseRequestFailedErr *samlprotocol.ParseRequestFailedError
			if errors.As(err, &parseRequestFailedErr) {
				err = NewSAMLErrorResult(err,
					samlprotocol.NewRequestDeniedErrorResponse(
						h.Clock.NowUTC(),
						h.SAMLService.IdpEntityID(),
						"failed to parse SAMLRequest",
						parseRequestFailedErr.GetDetailElements(),
					),
				)
			}
		}
	}()
	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		r, parseErr := samlbinding.SAMLBindingHTTPRedirectParse(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		logoutRequest, err = samlprotocol.ParseLogoutRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed LogoutRequest",
				Cause:  err,
			}
			return nil, nil, "", err
		}
		parseResult = r
		relayState = r.RelayState
	case "POST":
		// HTTP-POST binding
		r, parseErr := samlbinding.SAMLBindingHTTPPostParse(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		logoutRequest, err = samlprotocol.ParseLogoutRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed LogoutRequest",
				Cause:  err,
			}
			return nil, nil, "", err
		}
		parseResult = r
		relayState = r.RelayState
	default:
		// panic because it should not happen if ConfigureLogoutRoute is correct
		panic("unexpected method")
	}
	return parseResult, logoutRequest, relayState, nil
}

func (h *LogoutHandler) verifySignature(
	sp *config.SAMLServiceProviderConfig,
	parseResult samlbinding.SAMLBindingParseResult,
) (err error) {
	defer func() {
		// Transform known errors
		if err != nil {
			var invalidSignatureErr *samlprotocol.InvalidSignatureError
			if errors.As(err, &invalidSignatureErr) {
				err = NewSAMLErrorResult(err,
					samlprotocol.NewRequestDeniedErrorResponse(
						h.Clock.NowUTC(),
						h.SAMLService.IdpEntityID(),
						"invalid signature",
						invalidSignatureErr.GetDetailElements(),
					),
				)
			}
		}
	}()
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			parseResult.SAMLRequest,
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			return err
		}
	case *samlbinding.SAMLBindingHTTPPostParseResult:
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLRequestXML)
		if err != nil {
			return err
		}
	default:
		panic("unexpected parse result type")
	}
	return nil
}

func (h *LogoutHandler) invalidateSession(
	rw http.ResponseWriter,
	sp *config.SAMLServiceProviderConfig,
	sid string,
) (
	userID string,
	affectedServiceProviderIDs setutil.Set[string],
	err error,
) {
	_, sessionID, ok := oidc.DecodeSID(sid)
	if ok {
		s, err := h.SessionManager.Get(sessionID)
		if err != nil {
			if errors.Is(err, session.ErrSessionNotFound) {
				// If the session does not exist, simply ignore it
				return "", nil, nil
			} else {
				return "", nil, err
			}
		}
		userID := s.GetAuthenticationInfo().UserID
		invalidatedSessions, err := h.SessionManager.Logout(s, rw)
		if err != nil {
			return "", nil, err
		}
		for _, s := range invalidatedSessions {
			affectedServiceProviderIDs = affectedServiceProviderIDs.Merge(s.GetParticipatedSAMLServiceProviderIDs())
		}
		// Exclude the current logging out service provider
		affectedServiceProviderIDs.Delete(sp.GetID())
		return userID, affectedServiceProviderIDs, nil
	}
	return "", setutil.Set[string]{}, nil
}

func (h *LogoutHandler) handleSLORequest(
	rw http.ResponseWriter,
	r *http.Request,
	sp *config.SAMLServiceProviderConfig,
	responseBinding samlprotocol.SAMLBinding,
	callbackURL string,
) (relayState string, result logoutResult, err error) {
	var parseResult samlbinding.SAMLBindingParseResult
	var logoutRequest *samlprotocol.LogoutRequest

	// Get data with corresponding binding
	parseResult, logoutRequest, relayState, err = h.parseSLORequest(r)
	if err != nil {
		return relayState, nil, err
	}

	// Verify the signature
	err = h.verifySignature(sp, parseResult)
	if err != nil {
		return relayState, nil, err
	}

	var affectedServiceProviderIDs setutil.Set[string]
	var sid string
	var userID string
	if logoutRequest.SessionIndex != nil {
		sid = logoutRequest.SessionIndex.Value
		userID, affectedServiceProviderIDs, err = h.invalidateSession(rw, sp, sid)
		if err != nil {
			return relayState, nil, err
		}
		// Exclude the current logging out SP
		affectedServiceProviderIDs.Delete(sp.GetID())
	}

	if userID != "" && len(affectedServiceProviderIDs.Keys()) > 0 {
		sloSession, err := h.createSLOSession(
			sid,
			userID,
			logoutRequest,
			callbackURL,
			responseBinding,
			relayState,
			affectedServiceProviderIDs,
		)
		if err != nil {
			return relayState, nil, err
		}

		return relayState, &logoutRemainingSPsResult{
			sloSession: sloSession,
		}, nil
	}

	logoutResponse, err := h.SAMLService.IssueLogoutResponse(
		callbackURL,
		sp.GetID(),
		logoutRequest,
		false,
	)
	if err != nil {
		return relayState, nil, err
	}

	return relayState, &logoutCompleteResult{
		response: logoutResponse.Element(),
	}, nil
}

func (s *LogoutHandler) createSLOSession(
	sid string,
	userID string,
	request *samlprotocol.LogoutRequest,
	callbackURL string,
	responseBinding samlprotocol.SAMLBinding,
	relayState string,
	pendingLogoutServiceProviderIDs setutil.Set[string],
) (*samlslosession.SAMLSLOSession, error) {
	requestXML := string(request.ToXMLBytes())
	sloSessionEntry := &samlslosession.SAMLSLOSessionEntry{
		PendingLogoutServiceProviderIDs: pendingLogoutServiceProviderIDs,
		LogoutRequestXML:                requestXML,
		ResponseBinding:                 responseBinding,
		CallbackURL:                     callbackURL,
		RelayState:                      relayState,
		SID:                             sid,
		UserID:                          userID,
	}
	sloSession := samlslosession.NewSAMLSLOSession(sloSessionEntry)
	err := s.SAMLSLOSessionService.Save(sloSession)
	if err != nil {
		return nil, err
	}
	return sloSession, err
}

func (h *LogoutHandler) writeResponse(
	rw http.ResponseWriter, r *http.Request,
	responseEl *etree.Element,
	responseBinding samlprotocol.SAMLBinding,
	callbackURL string,
	relayState string,
) {
	switch responseBinding {
	case samlprotocol.SAMLBindingHTTPPost:
		err := h.BindingHTTPPostWriter.WriteResponse(rw, r,
			callbackURL,
			responseEl,
			relayState,
		)
		if err != nil {
			panic(err)
		}
	case samlprotocol.SAMLBindingHTTPRedirect:
		err := h.BindingHTTPRedirectWriter.WriteResponse(rw, r,
			callbackURL,
			responseEl,
			relayState,
		)
		if err != nil {
			panic(err)
		}
	}
}

func (h *LogoutHandler) handleError(
	rw http.ResponseWriter,
	r *http.Request,
	responseBinding samlprotocol.SAMLBinding,
	callbackURL string,
	relayState string,
	err error,
) {
	now := h.Clock.NowUTC()
	var samlErrResult *SAMLErrorResult
	var response samlprotocol.Respondable
	if errors.As(err, &samlErrResult) {
		h.Logger.WithError(samlErrResult.Cause).Warnln("saml logout failed with expected error")
		response = samlErrResult.Response
	} else {
		h.Logger.WithError(err).Error("unexpected error")
		response = samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID())
	}
	h.writeResponse(rw, r, response.Element(), responseBinding, callbackURL, relayState)
}

func (h *LogoutHandler) sendSLORequest(
	rw http.ResponseWriter,
	r *http.Request,
	sloSession *samlslosession.SAMLSLOSession,
	sp *config.SAMLServiceProviderConfig,
) error {
	logoutRequest, err := h.SAMLService.IssueLogoutRequest(
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
		err = h.BindingHTTPPostWriter.WriteRequest(rw, r, callbackURL, logoutRequestEl, sloSession.ID)
		if err != nil {
			return err
		}
	case samlprotocol.SAMLBindingHTTPRedirect:
		err = h.BindingHTTPRedirectWriter.WriteRequest(rw, r, callbackURL, logoutRequestEl, sloSession.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
