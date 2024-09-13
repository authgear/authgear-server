package saml

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/saml2/login/:service_provider_id")
}

type LoginHandlerLogger struct{ *log.Logger }

func NewLoginHandlerLogger(lf *log.Factory) *LoginHandlerLogger {
	return &LoginHandlerLogger{lf.New("saml-login-handler")}
}

type LoginHandler struct {
	Logger             *LoginHandlerLogger
	Clock              clock.Clock
	Database           *appdb.Handle
	SAMLConfig         *config.SAMLConfig
	SAMLService        HandlerSAMLService
	SAMLSessionService SAMLSessionService
	SAMLUIService      SAMLUIService

	UserFacade SAMLUserFacade

	LoginResultHandler LoginResultHandler
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	now := h.Clock.NowUTC()
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	sp, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	callbackURL := sp.DefaultAcsURL()
	issuer := h.SAMLService.IdpEntityID()
	var relayState string

	defer func() {
		if err := recover(); err != nil {
			h.handleUnknownError(rw, r, callbackURL, relayState, err)
		}
	}()

	var err error
	var parseResult samlbinding.SAMLBindingParseResult
	var authnRequest *samlprotocol.AuthnRequest

	// Get data with corresponding binding
	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		r, err := samlbinding.SAMLBindingHTTPRedirectParse(r)
		if err != nil {
			break
		}
		authnRequest, err = samlprotocol.ParseAuthnRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlerror.ParseRequestFailedError{
				Reason: "malformed AuthnRequest",
				Cause:  err,
			}
			break
		}
		parseResult = r
		relayState = r.RelayState
	case "POST":
		// HTTP-POST binding
		r, err := samlbinding.SAMLBindingHTTPPostParse(r)
		if err != nil {
			break
		}
		authnRequest, err = samlprotocol.ParseAuthnRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlerror.ParseRequestFailedError{
				Reason: "malformed AuthnRequest",
				Cause:  err,
			}
			break
		}
		parseResult = r
		relayState = r.RelayState
	default:
		panic(fmt.Errorf("unexpected method %s", r.Method))
	}

	if err != nil {
		if errors.Is(err, samlbinding.ErrNoRequest) {
			// This is a IdP-initated flow
			result := h.handleIdpInitiated(
				r.Context(),
				sp,
				callbackURL,
			)
			h.writeResult(rw, r, result)
			return
		}

		var parseRequestFailedErr *samlerror.ParseRequestFailedError
		if errors.As(err, &parseRequestFailedErr) {
			errResponse := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						"failed to parse SAMLRequest",
						parseRequestFailedErr.GetDetailElements(),
					),
				},
			)
			h.writeResult(rw, r, errResponse)
			return
		}
		panic(err)
	}

	// Verify the signature
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			parseResult.SAMLRequest,
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			break
		}
	case *samlbinding.SAMLBindingHTTPPostParseResult:
		relayState = parseResult.RelayState
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLRequestXML)
		if err != nil {
			break
		}
	default:
		panic("unexpected parse result type")
	}

	if err != nil {
		var invalidSignatureErr *samlerror.InvalidSignatureError
		if errors.As(err, &invalidSignatureErr) {
			errResponse := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						"invalid signature",
						invalidSignatureErr.GetDetailElements(),
					),
					RelayState: relayState,
				},
			)
			h.writeResult(rw, r, errResponse)
			return
		}
		panic(err)
	}

	// Validate the AuthnRequest
	err = h.SAMLService.ValidateAuthnRequest(sp.GetID(), authnRequest)
	if err != nil {
		var invalidRequestErr *samlerror.InvalidRequestError
		if errors.As(err, &invalidRequestErr) {
			errResponse := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						fmt.Sprintf("invalid AuthnRequest: %s", invalidRequestErr.Reason),
						invalidRequestErr.GetDetailElements(),
					),
					RelayState: relayState,
				},
			)
			h.writeResult(rw, r, errResponse)
			return
		}
		panic(err)
	}

	if authnRequest.AssertionConsumerServiceURL != "" {
		callbackURL = authnRequest.AssertionConsumerServiceURL
	}

	result := h.startSSOFlow(
		r.Context(),
		sp,
		string(authnRequest.ToXMLBytes()),
		callbackURL,
		relayState,
	)
	h.writeResult(rw, r, result)
}

func (h *LoginHandler) handleUnknownError(
	rw http.ResponseWriter, r *http.Request,
	callbackURL string,
	relayState string,
	err any,
) {
	now := h.Logger.Time.UTC()
	e := panicutil.MakeError(err)

	result := samlprotocolhttp.NewUnexpectedSAMLErrorResult(e,
		samlprotocolhttp.SAMLResult{
			CallbackURL: callbackURL,
			Binding:     samlprotocol.SAMLBindingHTTPPost,
			Response:    samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
			RelayState:  relayState,
		},
	)
	h.writeResult(rw, r, result)
}

func (h *LoginHandler) writeResult(
	rw http.ResponseWriter, r *http.Request,
	result httputil.Result,
) {
	switch result := result.(type) {
	case *samlprotocolhttp.SAMLErrorResult:
		if result.IsUnexpected {
			h.Logger.WithError(result.Cause).Error("unexpected error")
		} else {
			h.Logger.WithError(result).Warnln("saml login failed with expected error")
		}
	}
	result.WriteResponse(rw, r)
}

func (h *LoginHandler) handleIdpInitiated(
	ctx context.Context,
	sp *config.SAMLServiceProviderConfig,
	callbackURL string,
) httputil.Result {
	return h.startSSOFlow(
		ctx,
		sp,
		"",
		callbackURL,
		"",
	)
}

func (h *LoginHandler) startSSOFlow(
	ctx context.Context,
	sp *config.SAMLServiceProviderConfig,
	authnRequestXML string,
	callbackURL string,
	relayState string,
) httputil.Result {
	now := h.Clock.NowUTC()
	issuer := h.SAMLService.IdpEntityID()

	samlSessionEntry := &samlsession.SAMLSessionEntry{
		ServiceProviderID: sp.GetID(),
		AuthnRequestXML:   authnRequestXML,
		CallbackURL:       callbackURL,
		RelayState:        relayState,
	}

	uiInfo, showUI, err := h.SAMLUIService.ResolveUIInfo(sp, samlSessionEntry)
	if err != nil {
		var invalidRequestErr *samlerror.InvalidRequestError
		if errors.As(err, &invalidRequestErr) {
			return samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						fmt.Sprintf("invalid SAMLRequest: %s", invalidRequestErr.Reason),
						invalidRequestErr.GetDetailElements(),
					),
					RelayState: relayState,
				},
			)
		}
		panic(err)
	}

	var loginHint *oauth.LoginHint
	l, err := oauth.ParseLoginHint(uiInfo.LoginHint)
	if err == nil {
		loginHint = l
	}

	if !showUI {
		// If IsPassive=true, no ui should be displayed.
		// Authenticate by existing session or error.
		var resolvedSession session.ResolvedSession
		if s := session.GetSession(ctx); s != nil {
			resolvedSession = s
		}
		// Ignore any session that is not allow to be used here
		if !oauth.ContainsAllScopes(oauth.SessionScopes(resolvedSession), []string{oauth.PreAuthenticatedURLScope}) {
			resolvedSession = nil
		}

		// Ignore any session that does not match login_hint
		err = h.Database.WithTx(func() error {
			if loginHint != nil && resolvedSession != nil {
				hintUserIDs, err := h.UserFacade.GetUserIDsByLoginHint(loginHint)
				if err != nil {
					return err
				}
				hintUserIDsSet := setutil.NewStringSetFromSlice(hintUserIDs)
				if !hintUserIDsSet.Has(resolvedSession.GetAuthenticationInfo().UserID) {
					resolvedSession = nil
				}
			}
			return nil
		})
		if err != nil {
			return samlprotocolhttp.NewUnexpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response:    samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
					RelayState:  relayState,
				},
			)
		}

		if resolvedSession == nil {
			// No session, return NoPassive error.
			err := fmt.Errorf("no session but IsPassive=true")
			result := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewNoPassiveErrorResponse(
						now,
						issuer,
					),
					RelayState: relayState,
				},
			)
			return result
		} else {
			// Else, authenticate with the existing session.
			authInfo := resolvedSession.CreateNewAuthenticationInfoByThisSession()
			result := h.LoginResultHandler.handleLoginResult(&authInfo, samlSessionEntry)
			return result
		}

	}

	samlSession := samlsession.NewSAMLSession(samlSessionEntry, uiInfo)
	err = h.SAMLSessionService.Save(samlSession)
	if err != nil {
		panic(err)
	}

	endpoint, err := h.SAMLUIService.BuildAuthenticationURL(samlSession)
	if err != nil {
		panic(err)
	}

	result := &httputil.ResultRedirect{
		URL: endpoint.String(),
	}

	return result
}
