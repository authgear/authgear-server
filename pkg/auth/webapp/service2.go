package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type UIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SessionStore interface {
	Get(id string) (*Session, error)
	Create(session *Session) (err error)
	Update(session *Session) (err error)
	Delete(id string) (err error)
}

type GraphService interface {
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	Get(instanceID string) (*interaction.Graph, error)
	DryRun(contextValues interaction.ContextValues, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(contextValues interaction.ContextValues, graph *interaction.Graph) error
	Accept(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (*interaction.Graph, []interaction.Edge, error)
}

type CookiesGetter interface {
	GetCookies() []*http.Cookie
}

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("webapp-service")}
}

type Service2 struct {
	Logger               ServiceLogger
	Request              *http.Request
	Sessions             SessionStore
	SessionCookie        SessionCookieDef
	SignedUpCookie       SignedUpCookieDef
	MFADeviceTokenCookie mfa.CookieDef
	ErrorService         *ErrorService
	Cookies              CookieManager
	OAuthConfig          *config.OAuthConfig
	UIConfig             *config.UIConfig
	TrustProxy           config.TrustProxy
	UIInfoResolver       UIInfoResolver
	OAuthClientResolver  OAuthClientResolver

	Graph GraphService
}

func (s *Service2) CreateSession(session *Session, redirectURI string) (*Result, error) {
	if err := s.Sessions.Create(session); err != nil {
		return nil, err
	}
	result := &Result{
		RedirectURI: redirectURI,
		Cookies:     []*http.Cookie{s.Cookies.ValueCookie(s.SessionCookie.Def, session.ID)},
	}
	return result, nil
}

func (s *Service2) GetSession(id string) (*Session, error) {
	return s.Sessions.Get(id)
}

func (s *Service2) UpdateSession(session *Session) error {
	return s.Sessions.Update(session)
}

func (s *Service2) DeleteSession(sessionID string) error {
	return s.Sessions.Delete(sessionID)
}

// PeekUncommittedChanges runs fn with the effects of the graph fully applied.
// This is useful if fn needs the effects of the graph visible to it.
func (s *Service2) PeekUncommittedChanges(session *Session, fn func(graph *interaction.Graph) error) error {
	graph, err := s.Graph.Get(session.CurrentStep().GraphID)
	if err != nil {
		return ErrInvalidSession
	}

	err = s.Graph.DryRun(interaction.ContextValues{
		WebSessionID:   session.ID,
		OAuthSessionID: session.OAuthSessionID,
	}, func(ctx *interaction.Context) (*interaction.Graph, error) {
		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}
		err = fn(graph)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service2) Get(session *Session) (*interaction.Graph, error) {
	graph, err := s.Graph.Get(session.CurrentStep().GraphID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	err = s.Graph.DryRun(interaction.ContextValues{
		WebSessionID:   session.ID,
		OAuthSessionID: session.OAuthSessionID,
	}, func(ctx *interaction.Context) (*interaction.Graph, error) {
		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return graph, nil
}

func (s *Service2) GetWithIntent(session *Session, intent interaction.Intent) (*interaction.Graph, error) {
	var graph *interaction.Graph
	err := s.Graph.DryRun(interaction.ContextValues{
		WebSessionID:   session.ID,
		OAuthSessionID: session.OAuthSessionID,
	}, func(ctx *interaction.Context) (*interaction.Graph, error) {
		g, err := s.Graph.NewGraph(ctx, intent)
		if err != nil {
			return nil, err
		}

		graph = g
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return graph, nil
}

func (s *Service2) PostWithIntent(
	session *Session,
	intent interaction.Intent,
	inputFn func() (interface{}, error),
) (result *Result, err error) {
	return s.doPost(session, inputFn, true, func(ctx *interaction.Context) (*interaction.Graph, error) {
		return s.Graph.NewGraph(ctx, intent)
	})
}

func (s *Service2) PostWithInput(
	session *Session,
	inputFn func() (interface{}, error),
) (result *Result, err error) {
	return s.doPost(session, inputFn, false, func(ctx *interaction.Context) (*interaction.Graph, error) {
		return s.Graph.Get(session.CurrentStep().GraphID)
	})
}

// nolint: gocognit
func (s *Service2) doPost(
	session *Session,
	inputFn func() (interface{}, error),
	isNewGraph bool,
	graphFn func(ctx *interaction.Context) (*interaction.Graph, error),
) (*Result, error) {
	result := &Result{}
	deviceTokenAttempted := false

	var graph *interaction.Graph
	var isFinished bool
	var err error
	for {
		var edges []interaction.Edge
		graph, edges, err = s.runGraph(session, result, inputFn, graphFn)
		isFinished = len(edges) == 0 && err == nil
		if err != nil || isFinished {
			break
		}

		kind := deriveSessionStepKind(graph)
		session.Steps = append(
			session.Steps,
			NewSessionStep(deriveSessionStepKind(graph), graph.InstanceID),
		)

		inputFn = nil
		// Select default for steps with choice
		switch kind {
		case SessionStepAuthenticate:
			authDeviceToken := ""
			if deviceTokenCookie, err := s.Cookies.GetCookie(s.Request, s.MFADeviceTokenCookie.Def); err == nil {
				for _, edge := range edges {
					if _, ok := edge.(*nodes.EdgeUseDeviceToken); ok {
						authDeviceToken = deviceTokenCookie.Value
						break
					}
				}
			}
			if len(authDeviceToken) > 0 && !deviceTokenAttempted {
				inputFn = func() (interface{}, error) {
					return &inputAuthDeviceToken{
						DeviceToken: authDeviceToken,
					}, nil
				}
				deviceTokenAttempted = true
				break
			}

			switch defaultEdge := edges[0].(type) {
			case *nodes.EdgeConsumeRecoveryCode:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepEnterRecoveryCode,
					graph.InstanceID,
				))
			case *nodes.EdgeAuthenticationPassword:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepEnterPassword,
					graph.InstanceID,
				))
			case *nodes.EdgeAuthenticationPasskey:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepUsePasskey,
					graph.InstanceID,
				))
			case *nodes.EdgeAuthenticationTOTP:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepEnterTOTP,
					graph.InstanceID,
				))
			case *nodes.EdgeAuthenticationOOBTrigger:
				inputFn = func() (input interface{}, err error) {
					input = &inputTriggerOOB{
						AuthenticatorType:  string(defaultEdge.OOBAuthenticatorType),
						AuthenticatorIndex: 0,
					}
					return
				}
			case *nodes.EdgeAuthenticationWhatsappTrigger:
				inputFn = func() (input interface{}, err error) {
					input = &inputTriggerWhatsapp{
						AuthenticatorIndex: 0,
					}
					return
				}
			case *nodes.EdgeAuthenticationLoginLinkTrigger:
				inputFn = func() (input interface{}, err error) {
					input = &inputTriggerLoginLink{
						AuthenticatorIndex: 0,
					}
					return
				}
			default:
				panic(fmt.Errorf("webapp: unexpected edge: %T", defaultEdge))
			}

		case SessionStepCreateAuthenticator:
			switch defaultEdge := edges[0].(type) {
			case *nodes.EdgeCreateAuthenticatorPassword:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepCreatePassword,
					graph.InstanceID,
				))
			case *nodes.EdgeCreateAuthenticatorPasskey:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepCreatePasskey,
					graph.InstanceID,
				))
			case *nodes.EdgeCreateAuthenticatorOOBSetup:
				if defaultEdge.Stage == authn.AuthenticationStagePrimary {
					inputFn = func() (interface{}, error) {
						return &inputSelectOOB{}, nil
					}
				} else {
					var stepKind SessionStepKind
					switch defaultEdge.AuthenticatorType() {
					case model.AuthenticatorTypeOOBEmail:
						stepKind = SessionStepSetupOOBOTPEmail
					case model.AuthenticatorTypeOOBSMS:
						stepKind = SessionStepSetupOOBOTPSMS
					default:
						panic(fmt.Errorf("webapp: unexpected authenticator type in oob edge: %s", defaultEdge.AuthenticatorType()))
					}
					session.Steps = append(session.Steps, NewSessionStep(
						stepKind,
						graph.InstanceID,
					))
				}
			case *nodes.EdgeCreateAuthenticatorTOTPSetup:
				inputFn = func() (interface{}, error) {
					return &inputSelectTOTP{}, nil
				}
			case *nodes.EdgeCreateAuthenticatorWhatsappOTPSetup:
				if defaultEdge.Stage == authn.AuthenticationStagePrimary {
					inputFn = func() (interface{}, error) {
						return &inputSelectWhatsappOTP{}, nil
					}
				} else {
					session.Steps = append(session.Steps, NewSessionStep(
						SessionStepSetupWhatsappOTP,
						graph.InstanceID,
					))
				}
			case *nodes.EdgeCreateAuthenticatorLoginLinkOTPSetup:
				if defaultEdge.Stage == authn.AuthenticationStagePrimary {
					inputFn = func() (interface{}, error) {
						return &inputSelectLoginLinkOTP{}, nil
					}
				} else {
					session.Steps = append(session.Steps, NewSessionStep(
						SessionStepSetupLoginLinkOTP,
						graph.InstanceID,
					))
				}
			default:
				panic(fmt.Errorf("webapp: unexpected edge: %T", defaultEdge))
			}

		case SessionStepVerifyIdentityBegin:
			switch defaultEdge := edges[0].(type) {
			case *nodes.EdgeVerifyIdentity:
				inputFn = func() (interface{}, error) {
					return &inputSelectVerifyIdentityViaOOBOTP{}, nil
				}
			case *nodes.EdgeVerifyIdentityViaWhatsapp:
				inputFn = func() (interface{}, error) {
					return &inputSelectVerifyIdentityViaWhatsapp{}, nil
				}
			default:
				panic(fmt.Errorf("webapp: unexpected edge: %T", defaultEdge))
			}
		}

		if inputFn == nil {
			break
		}

		// Continue to feed new input.
		graphFn = func(ctx *interaction.Context) (*interaction.Graph, error) {
			return graph, nil
		}
	}

	err = s.afterPost(result, session, graph, err, isFinished, isNewGraph)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service2) runGraph(
	session *Session,
	result *Result,
	inputFn func() (interface{}, error),
	graphFn func(ctx *interaction.Context) (*interaction.Graph, error),
) (*interaction.Graph, []interaction.Edge, error) {
	var graph *interaction.Graph
	var edges []interaction.Edge
	interactionErr := s.Graph.DryRun(interaction.ContextValues{
		WebSessionID:   session.ID,
		OAuthSessionID: session.OAuthSessionID,
	}, func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		graph, err = graphFn(ctx)
		if err != nil {
			return nil, err
		}

		input, err := inputFn()
		if err != nil {
			return nil, err
		}

		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}

		var newGraph *interaction.Graph
		newGraph, edges, err = s.Graph.Accept(ctx, graph, input)

		var clearCookie *interaction.ErrClearCookie
		var inputRequired *interaction.ErrInputRequired
		if errors.As(err, &clearCookie) {
			result.Cookies = append(result.Cookies, clearCookie.Cookies...)
			graph = newGraph
			return newGraph, nil
		} else if errors.As(err, &inputRequired) {
			graph = newGraph
			return newGraph, nil
		} else if err != nil {
			return nil, err
		}

		graph = newGraph
		return newGraph, nil
	})

	return graph, edges, interactionErr
}

func (s *Service2) afterPost(
	result *Result,
	session *Session,
	graph *interaction.Graph,
	interactionErr error,
	isFinished bool,
	isNewGraph bool,
) error {
	if isFinished {
		// The graph finished. Apply its effect permanently.
		s.Logger.Debugf("interaction: commit graph")
		interactionErr = s.Graph.Run(interaction.ContextValues{
			WebSessionID:   session.ID,
			OAuthSessionID: session.OAuthSessionID,
		}, graph)
		// The interaction should not be finished, if there is error when
		// applying effects.
		isFinished = interactionErr == nil
	}

	// Populate cookies.
	if interactionErr != nil {
		if !apierrors.IsAPIError(interactionErr) {
			s.Logger.WithError(interactionErr).Error("interaction error")
		}
		errCookie, err := s.ErrorService.SetRecoverableError(s.Request, apierrors.AsAPIError(interactionErr))
		if err != nil {
			return err
		}
		result.IsInteractionErr = true
		result.Cookies = append(result.Cookies, errCookie)
	} else if isFinished {
		// Loop from start to end to collect cookies.
		// This iteration order allows newer node to overwrite cookies.
		for _, node := range graph.Nodes {
			if a, ok := node.(CookiesGetter); ok {
				result.Cookies = append(result.Cookies, a.GetCookies()...)
			}
		}
	}

	// Transition to redirect URI
	if isFinished {
		result.RemoveQueries = setutil.Set[string]{
			"x_step": struct{}{},
		}
		result.RedirectURI = s.deriveFinishRedirectURI(session, graph)
		switch graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			result.NavigationAction = "redirect"
			// Marked signed up in cookie after authorization.
			// When user visit auth ui root "/", redirect user to "/login" if
			// cookie exists
			result.Cookies = append(result.Cookies, s.Cookies.ValueCookie(s.SignedUpCookie.Def, "true"))
		default:
			// Use the default navigation action for any other intents.
			// That is, "advance" will be used.
			// This is more suitable because interactions started in settings will have finish URI within settings.
			// Keep using Turbo make sure "popstate" will be fired properly.
			// If "redirect" is used, "popstate" will NOT be fired.
			break
		}
	} else if interactionErr == nil {
		if a, ok := graph.CurrentNode().(interface{ GetRedirectURI() string }); ok {
			result.RedirectURI = a.GetRedirectURI()
		} else {
			result.RedirectURI = session.CurrentStep().URL().String()
		}
	} else {
		if a, ok := graph.CurrentNode().(interface{ GetErrorRedirectURI() string }); ok {
			result.RedirectURI = a.GetErrorRedirectURI()
		} else {
			u := url.URL{
				Path:     s.Request.URL.Path,
				RawQuery: s.Request.URL.RawQuery,
			}
			result.RedirectURI = u.String()
			result.NavigationAction = "replace"
		}
	}
	s.Logger.Debugf("interaction: redirect to %s", result.RedirectURI)

	// Collect extras
	session.Extra = collectExtras(graph.CurrentNode())

	// Persist/discard session
	if isFinished && !session.KeepAfterFinish {
		err := s.Sessions.Delete(session.ID)
		if err != nil {
			return err
		}
		result.Cookies = append(result.Cookies, s.Cookies.ClearCookie(s.SessionCookie.Def))
		// reset visitor id after completing interaction
		// visit signup / login page again after signup should treat as a new visit
		result.Cookies = append(result.Cookies, s.Cookies.ClearCookie(VisitorIDCookieDef))
	} else if isNewGraph {
		err := s.Sessions.Create(session)
		if err != nil {
			return err
		}
		result.Cookies = append(result.Cookies, s.Cookies.ValueCookie(s.SessionCookie.Def, session.ID))
	} else if interactionErr == nil {
		err := s.Sessions.Update(session)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service2) deriveFinishRedirectURI(session *Session, graph *interaction.Graph) (redirectURI string) {
	defer func() {
		if e, ok := graph.GetAuthenticationInfoEntry(); ok {
			redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(redirectURI, e)
		}
	}()

	switch graph.CurrentNode().(type) {
	case *nodes.NodeForgotPasswordEnd:
		redirectURI = "/flows/forgot_password/success"
		return
	case *nodes.NodeResetPasswordEnd:
		redirectURI = "/flows/reset_password/success"
		return
	}

	// 1. WebSession's Redirect URL (e.g. authorization endpoint, settings page)
	// 2. Obtain redirect_uri which is from the same origin by calling GetRedirectURI
	// 3. DerivePostLoginRedirectURIFromRequest
	if session.RedirectURI != "" {
		redirectURI = session.RedirectURI
		return
	}

	postLoginRedirectURI := DerivePostLoginRedirectURIFromRequest(s.Request, s.OAuthClientResolver, s.UIConfig)
	redirectURI = GetRedirectURI(s.Request, bool(s.TrustProxy), postLoginRedirectURI)
	return
}

func deriveSessionStepKind(graph *interaction.Graph) SessionStepKind {
	switch currentNode := graph.CurrentNode().(type) {
	case *nodes.NodeUseIdentityOAuthProvider:
		return SessionStepOAuthRedirect
	case *nodes.NodeCreateIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindPromote:
				return SessionStepPromoteUser
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeAuthenticationBegin:
		return SessionStepAuthenticate
	case *nodes.NodeReauthenticationBegin:
		return SessionStepAuthenticate
	case *nodes.NodeCreateAuthenticatorBegin:
		return SessionStepCreateAuthenticator
	case *nodes.NodeAuthenticationOOBTrigger:
		switch currentNode.Authenticator.Type {
		case model.AuthenticatorTypeOOBEmail:
			return SessionStepEnterOOBOTPAuthnEmail
		case model.AuthenticatorTypeOOBSMS:
			return SessionStepEnterOOBOTPAuthnSMS
		default:
			panic(fmt.Errorf("webapp: unexpected oob authenticator type: %s", currentNode.Authenticator.Type))
		}
	case *nodes.NodeAuthenticationWhatsappTrigger:
		return SessionStepVerifyWhatsappOTPAuthn
	case *nodes.NodeAuthenticationLoginLinkTrigger:
		return SessionStepVerifyLoginLinkOTPAuthn
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		switch currentNode.Authenticator.Type {
		case model.AuthenticatorTypeOOBEmail:
			return SessionStepEnterOOBOTPSetupEmail
		case model.AuthenticatorTypeOOBSMS:
			return SessionStepEnterOOBOTPSetupSMS
		default:
			panic(fmt.Errorf("webapp: unexpected oob authenticator type: %s", currentNode.Authenticator.Type))
		}
	case *nodes.NodeCreateAuthenticatorWhatsappOTPSetup:
		return SessionStepVerifyWhatsappOTPSetup
	case *nodes.NodeCreateAuthenticatorLoginLinkOTPSetup:
		return SessionStepVerifyLoginLinkOTPSetup
	case *nodes.NodeCreateAuthenticatorTOTPSetup:
		return SessionStepSetupTOTP
	case *nodes.NodeGenerateRecoveryCodeBegin:
		return SessionStepSetupRecoveryCode
	case *nodes.NodeEnsureVerificationBegin:
		return SessionStepVerifyIdentityBegin
	case *nodes.NodeVerifyIdentity:
		return SessionStepVerifyIdentityViaOOBOTP
	case *nodes.NodeVerifyIdentityViaWhatsapp:
		return SessionStepVerifyIdentityViaWhatsapp
	case *nodes.NodeValidateUser:
		return SessionStepAccountStatus
	case *nodes.NodeChangePasswordBegin:
		switch currentNode.Stage {
		case authn.AuthenticationStagePrimary:
			return SessionStepChangePrimaryPassword
		case authn.AuthenticationStageSecondary:
			return SessionStepChangeSecondaryPassword
		default:
			panic(fmt.Errorf("webapp: unexpected authentication stage: %s", currentNode.Stage))
		}
	case *nodes.NodePromptCreatePasskeyBegin:
		return SessionStepPromptCreatePasskey
	case *nodes.NodeConfirmTerminateOtherSessionsBegin:
		return SessionStepConfirmTerminateOtherSessions
	default:
		panic(fmt.Errorf("webapp: unexpected node: %T", graph.CurrentNode()))
	}
}

func collectExtras(node interaction.Node) map[string]interface{} {
	switch node := node.(type) {
	case *nodes.NodeForgotPasswordEnd:
		return map[string]interface{}{
			"login_id": node.LoginID,
		}
	case *nodes.NodeEnsureVerificationEnd:
		return map[string]interface{}{
			"display_id": node.Identity.DisplayID(),
		}
	case *nodes.NodeDoVerifyIdentity:
		return map[string]interface{}{
			"display_id": node.Identity.DisplayID(),
		}
	default:
		return nil
	}
}
