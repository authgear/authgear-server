package saml

import (
	"errors"
	"net/http"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlslosession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
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

// logoutCompleteResult means the whole logout process is completed,
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
	Database              *appdb.Handle
	SAMLConfig            *config.SAMLConfig
	SAMLService           HandlerSAMLService
	SessionManager        SessionManager
	SAMLSLOSessionService SAMLSLOSessionService
	SAMLSLOService        SAMLSLOService
	Endpoints             Endpoints

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

	err := h.Database.WithTx(func() error {
		h.handle(rw, r, sp)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (h *LogoutHandler) handle(
	rw http.ResponseWriter,
	r *http.Request,
	sp *config.SAMLServiceProviderConfig) {

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
		if errors.Is(err, samlbinding.ErrNoRequest) {
			// No request found, try to parse it as response
			result, err = h.handleSLOResponse(rw, r, sp)
			if err != nil {
				if errors.Is(err, samlbinding.ErrNoResponse) {
					// No request nor response, Redirect to /logout for IdP-initiated logout
					http.Redirect(rw, r, h.Endpoints.LogoutEndpointURL().String(), http.StatusFound)
					return
				}
				// panic here because we are handling a response, so no need to return the error as a response.
				panic(err)
			}
		} else {
			h.handleError(rw, r, responseBinding, callbackURL, relayState, err)
			return
		}
	}

	switch result := result.(type) {
	case *logoutCompleteResult:
		// Finish the logout with a response
		h.writeResponse(rw, r, result.response, responseBinding, callbackURL, relayState)
		return
	case *logoutRemainingSPsResult:
		// Logout all remaining participants before finish
		h.doLogoutRemainingSPs(rw, r, result)
		return
	}
}

func (h *LogoutHandler) doLogoutRemainingSPs(
	rw http.ResponseWriter,
	r *http.Request,
	result *logoutRemainingSPsResult,
) {
	var err error
	sloSession := result.sloSession
	for _, spID := range result.sloSession.Entry.PendingLogoutServiceProviderIDs {
		sp, ok := h.SAMLConfig.ResolveProvider(spID)
		if ok && sp.SLOEnabled {
			err = h.SAMLSLOService.SendSLORequest(
				rw, r,
				result.sloSession,
				sp,
			)
			if err != nil {
				// For some reason it failed
				// Skip this SP and send request to the next one
				h.Logger.WithError(err).Error("failed to send logout request")
				sloSession.Entry.IsPartialLogout = true
				err = h.SAMLSLOSessionService.Save(sloSession)
				if err != nil {
					h.handleError(rw, r,
						result.sloSession.Entry.ResponseBinding,
						result.sloSession.Entry.CallbackURL,
						result.sloSession.Entry.RelayState,
						err,
					)
					return
				}
				continue
			}
			return
		}
	}
	// None of the SPs has slo enabled, end the logout immediately
	if logoutRequest, ok := result.sloSession.Entry.LogoutRequest(); ok {
		logoutResponse, err := h.SAMLService.IssueLogoutResponse(
			result.sloSession.Entry.CallbackURL,
			logoutRequest,
			sloSession.Entry.IsPartialLogout,
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
	} else if result.sloSession.Entry.PostLogoutRedirectURI != "" {
		// This is not a logout triggered by SP, redirect to post logout url
		http.Redirect(rw, r, result.sloSession.Entry.PostLogoutRedirectURI, http.StatusFound)
		return
	} else {
		panic("LogoutRequest and PostLogoutRedirectURI are empty, cannot determine the next action")
	}
}

func (h *LogoutHandler) parseSLORequest(
	r *http.Request,
) (
	parseResult samlbinding.SAMLBindingParseReqeustResult,
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
		r, parseErr := samlbinding.SAMLBindingHTTPRedirectParseRequest(r)
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
		r, parseErr := samlbinding.SAMLBindingHTTPPostParseRequest(r)
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

func (h *LogoutHandler) parseSLOResponse(
	r *http.Request,
) (
	parseResult samlbinding.SAMLBindingParseResponseResult,
	logoutResponse *samlprotocol.LogoutResponse,
	relayState string,
	err error,
) {
	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		r, parseErr := samlbinding.SAMLBindingHTTPRedirectParseResponse(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		logoutResponse, err = samlprotocol.ParseLogoutResponse([]byte(r.SAMLResponseXML))
		if err != nil {
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed LogoutResponse",
				Cause:  err,
			}
			return nil, nil, "", err
		}
		parseResult = r
		relayState = r.RelayState
	case "POST":
		// HTTP-POST binding
		r, parseErr := samlbinding.SAMLBindingHTTPPostParseResponse(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		logoutResponse, err = samlprotocol.ParseLogoutResponse([]byte(r.SAMLResponseXML))
		if err != nil {
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed LogoutResponse",
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
	return parseResult, logoutResponse, relayState, nil
}

func (h *LogoutHandler) verifyRequestSignature(
	sp *config.SAMLServiceProviderConfig,
	parseResult samlbinding.SAMLBindingParseReqeustResult,
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
	case *samlbinding.SAMLBindingHTTPRedirectParseRequestResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			&saml.SAMLElementSigned{
				SAMLRequest: parseResult.SAMLRequest,
			},
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			return err
		}
	case *samlbinding.SAMLBindingHTTPPostParseRequestResult:
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLRequestXML)
		if err != nil {
			return err
		}
	default:
		panic("unexpected parse result type")
	}
	return nil
}

func (h *LogoutHandler) verifyResponseSignature(
	sp *config.SAMLServiceProviderConfig,
	parseResult samlbinding.SAMLBindingParseResponseResult,
) (err error) {
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseResponseResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			&saml.SAMLElementSigned{
				SAMLResponse: parseResult.SAMLResponse,
			},
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			return err
		}
	case *samlbinding.SAMLBindingHTTPPostParseResponseResult:
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLResponseXML)
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
			affectedServiceProviderIDs = affectedServiceProviderIDs.Merge(s.GetParticipatedSAMLServiceProviderIDsSet())
		}
		// Exclude the current logging out service provider
		affectedServiceProviderIDs.Delete(sp.GetID())
		return userID, affectedServiceProviderIDs, nil
	}
	return "", setutil.Set[string]{}, nil
}

func (h *LogoutHandler) handleSLOResponse(
	rw http.ResponseWriter,
	r *http.Request,
	sp *config.SAMLServiceProviderConfig,
) (result logoutResult, err error) {
	var parseResult samlbinding.SAMLBindingParseResponseResult
	var logoutRequest *samlprotocol.LogoutResponse
	var relayState string

	// Get data with corresponding binding
	parseResult, logoutRequest, relayState, err = h.parseSLOResponse(r)
	if err != nil {
		return nil, err
	}

	// Verify the signature
	err = h.verifyResponseSignature(sp, parseResult)
	if err != nil {
		return nil, err
	}
	sloSessionID := relayState
	sloSession, err := h.SAMLSLOSessionService.Get(sloSessionID)
	if err != nil {
		// We do not check if it is ErrNotFound,
		// because it is unexpected that we receive an logout response without a slo session
		return nil, err
	}
	if logoutRequest.Status.StatusCode.Value != samlprotocol.StatusSuccess {
		// At least one logout is failed, return a correct status to indicate it
		sloSession.Entry.IsPartialLogout = true
	}
	// Remove the current SP id from the pending sp ids
	newPendingLogoutServiceProviderIDsSet := setutil.NewSetFromSlice(sloSession.Entry.PendingLogoutServiceProviderIDs, setutil.Identity)
	newPendingLogoutServiceProviderIDsSet.Delete(sp.GetID())
	sloSession.Entry.PendingLogoutServiceProviderIDs = newPendingLogoutServiceProviderIDsSet.Keys()

	err = h.SAMLSLOSessionService.Save(sloSession)
	if err != nil {
		return nil, err
	}

	return &logoutRemainingSPsResult{
		sloSession: sloSession,
	}, nil
}

func (h *LogoutHandler) handleSLORequest(
	rw http.ResponseWriter,
	r *http.Request,
	sp *config.SAMLServiceProviderConfig,
	responseBinding samlprotocol.SAMLBinding,
	callbackURL string,
) (relayState string, result logoutResult, err error) {
	var parseResult samlbinding.SAMLBindingParseReqeustResult
	var logoutRequest *samlprotocol.LogoutRequest

	// Get data with corresponding binding
	parseResult, logoutRequest, relayState, err = h.parseSLORequest(r)
	if err != nil {
		return relayState, nil, err
	}

	// Verify the signature
	err = h.verifyRequestSignature(sp, parseResult)
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
		PendingLogoutServiceProviderIDs: pendingLogoutServiceProviderIDs.Keys(),
		LogoutRequestXML:                requestXML,
		ResponseBinding:                 responseBinding,
		CallbackURL:                     callbackURL,
		RelayState:                      relayState,
		SID:                             sid,
		UserID:                          userID,
		IsPartialLogout:                 false,
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
