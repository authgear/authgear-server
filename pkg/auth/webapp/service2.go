package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type SessionStore interface {
	Get(id string) (*Session, error)
	Create(session *Session) (err error)
	Update(session *Session) (err error)
	Delete(id string) (err error)
}

type GraphService interface {
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	Get(instanceID string) (*interaction.Graph, error)
	DryRun(webStateID string, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(webStateID string, graph *interaction.Graph) error
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
	ErrorCookie          *ErrorCookie
	CookieFactory        CookieFactory

	Graph GraphService
}

func (s *Service2) CreateSession(session *Session, redirectURI string) (*Result, error) {
	if err := s.Sessions.Create(session); err != nil {
		return nil, err
	}
	result := &Result{
		RedirectURI: redirectURI,
		Cookies:     []*http.Cookie{s.CookieFactory.ValueCookie(s.SessionCookie.Def, session.ID)},
	}
	return result, nil
}

func (s *Service2) GetSession(id string) (*Session, error) {
	return s.Sessions.Get(id)
}

func (s *Service2) UpdateSession(session *Session) error {
	return s.Sessions.Update(session)
}

func (s *Service2) Get(session *Session) (*interaction.Graph, error) {
	graph, err := s.Graph.Get(session.CurrentStep().GraphID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	err = s.Graph.DryRun(session.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
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
	err := s.Graph.DryRun(session.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
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
			if deviceTokenCookie, err := s.Request.Cookie(s.MFADeviceTokenCookie.Def.Name); err == nil {
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
			case *nodes.EdgeAuthenticationTOTP:
				session.Steps = append(session.Steps, NewSessionStep(
					SessionStepEnterTOTP,
					graph.InstanceID,
				))
			case *nodes.EdgeAuthenticationOOBTrigger:
				switch defaultEdge.OOBAuthenticatorType {
				case authn.AuthenticatorTypeOOBEmail:
					session.Steps = append(session.Steps, NewSessionStep(
						SessionStepSendOOBOTPAuthnEmail,
						graph.InstanceID,
					))
				case authn.AuthenticatorTypeOOBSMS:
					session.Steps = append(session.Steps, NewSessionStep(
						SessionStepSendOOBOTPAuthnSMS,
						graph.InstanceID,
					))
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
			case *nodes.EdgeCreateAuthenticatorOOBSetup:
				var stepKind SessionStepKind
				switch defaultEdge.AuthenticatorType() {
				case authn.AuthenticatorTypeOOBEmail:
					stepKind = SessionStepSetupOOBOTPEmail
				case authn.AuthenticatorTypeOOBSMS:
					stepKind = SessionStepSetupOOBOTPSMS
				default:
					panic(fmt.Errorf("webapp: unexpected authenticator type in oob edge: %s", defaultEdge.AuthenticatorType()))
				}
				session.Steps = append(session.Steps, NewSessionStep(
					stepKind,
					graph.InstanceID,
				))
			case *nodes.EdgeCreateAuthenticatorTOTPSetup:
				inputFn = func() (interface{}, error) {
					return &inputSelectTOTP{}, nil
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
	interactionErr := s.Graph.DryRun(session.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
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
		interactionErr = s.Graph.Run(session.ID, graph)
		// The interaction should not be finished, if there is error when
		// applying effects.
		isFinished = interactionErr == nil
	}

	// Populate cookies.
	if interactionErr != nil {
		if !apierrors.IsAPIError(interactionErr) {
			s.Logger.Errorf("interaction error: %v", interactionErr)
		}
		errCookie, err := s.ErrorCookie.SetError(apierrors.AsAPIError(interactionErr))
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
		result.RedirectURI = deriveFinishRedirectURI(session, graph)
		switch graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			result.NavigationAction = "redirect"
			// Marked signed up in cookie after authorization.
			// When user visit auth ui root "/", redirect user to "/login" if
			// cookie exists
			result.Cookies = append(result.Cookies, s.CookieFactory.ValueCookie(s.SignedUpCookie.Def, "true"))
		default:
			// Use the default navigation action for any other intents.
			// That is, "advance" will be used.
			// This is more suitable because interactions started in settings will have finish URI within settings.
			// Keep using Turbolinks make sure "popstate" will be fired properly.
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
		}
	}
	s.Logger.Debugf("interaction: redirect to" + result.RedirectURI)

	// Collect extras
	session.Extra = collectExtras(graph.CurrentNode())

	// Persist/discard session
	if isFinished && !session.KeepAfterFinish {
		err := s.Sessions.Delete(session.ID)
		if err != nil {
			return err
		}
		result.Cookies = append(result.Cookies, s.CookieFactory.ClearCookie(s.SessionCookie.Def))
	} else if isNewGraph {
		err := s.Sessions.Create(session)
		if err != nil {
			return err
		}
		result.Cookies = append(result.Cookies, s.CookieFactory.ValueCookie(s.SessionCookie.Def, session.ID))
	} else if interactionErr == nil {
		err := s.Sessions.Update(session)
		if err != nil {
			return err
		}
	}

	return nil
}

// nolint:gocyclo
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
	case *nodes.NodeCreateAuthenticatorBegin:
		return SessionStepCreateAuthenticator
	case *nodes.NodeAuthenticationOOBTrigger:
		switch currentNode.Authenticator.Type {
		case authn.AuthenticatorTypeOOBEmail:
			return SessionStepEnterOOBOTPAuthnEmail
		case authn.AuthenticatorTypeOOBSMS:
			return SessionStepEnterOOBOTPAuthnSMS
		default:
			panic(fmt.Errorf("webapp: unexpected oob authenticator type: %s", currentNode.Authenticator.Type))
		}
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		switch currentNode.Authenticator.Type {
		case authn.AuthenticatorTypeOOBEmail:
			return SessionStepEnterOOBOTPSetupEmail
		case authn.AuthenticatorTypeOOBSMS:
			return SessionStepEnterOOBOTPSetupSMS
		default:
			panic(fmt.Errorf("webapp: unexpected oob authenticator type: %s", currentNode.Authenticator.Type))
		}
	case *nodes.NodeCreateAuthenticatorTOTPSetup:
		return SessionStepSetupTOTP
	case *nodes.NodeGenerateRecoveryCodeBegin:
		return SessionStepSetupRecoveryCode
	case *nodes.NodeVerifyIdentity:
		return SessionStepVerifyIdentity
	case *nodes.NodeValidateUser:
		return SessionStepUserDisabled
	default:
		panic(fmt.Errorf("webapp: unexpected node: %T", graph.CurrentNode()))
	}
}

func deriveFinishRedirectURI(session *Session, graph *interaction.Graph) string {
	switch graph.CurrentNode().(type) {
	case *nodes.NodeForgotPasswordEnd:
		return "/forgot_password/success"
	case *nodes.NodeResetPasswordEnd:
		return "/reset_password/success"
	}
	return session.RedirectURI
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
