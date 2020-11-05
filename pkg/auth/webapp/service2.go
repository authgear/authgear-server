package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type SessionStore interface {
	Create(session *Session) (err error)
	Update(session *Session) (err error)
	Delete(id string) (err error)
}

type Service2 struct {
	Logger               ServiceLogger
	Request              *http.Request
	Sessions             SessionStore
	SessionCookie        SessionCookieDef
	MFADeviceTokenCookie mfa.CookieDef
	ErrorCookie          *ErrorCookie
	CookieFactory        CookieFactory

	Graph GraphService
}

func (s *Service2) UpdateSession(session *Session) error {
	return s.Sessions.Update(session)
}

func (s *Service2) rewindSessionHistory(session *Session, path string) *SessionStep {
	stepGraphID := s.Request.Form.Get("x_step")
	var curStep *SessionStep
	for i := len(session.Steps) - 1; i >= 0; i-- {
		step := session.Steps[i]
		if step.GraphID == stepGraphID {
			curStep = &step
			break
		}
	}

	if curStep == nil && len(session.Steps) > 0 {
		s := session.CurrentStep()
		curStep = &s
	}
	if curStep == nil || !curStep.Kind.MatchPath(path) {
		return nil
	}
	return curStep
}

func (s *Service2) Get(path string, session *Session) (*interaction.Graph, error) {
	step := s.rewindSessionHistory(session, path)
	if step == nil {
		return nil, ErrSessionStepMismatch
	}

	graph, err := s.Graph.Get(step.GraphID)
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
	path string,
	session *Session,
	inputFn func() (interface{}, error),
) (result *Result, err error) {
	step := s.rewindSessionHistory(session, path)
	if step == nil {
		return nil, ErrSessionStepMismatch
	}

	return s.doPost(session, inputFn, false, func(ctx *interaction.Context) (*interaction.Graph, error) {
		return s.Graph.Get(step.GraphID)
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
		graph, edges, err = s.runGraph(result, inputFn, graphFn)
		isFinished = len(edges) == 0 && err == nil
		if err != nil || isFinished {
			break
		}

		kind := deriveSessionStepKind(graph)
		session.Steps = append(session.Steps, SessionStep{
			Kind:    deriveSessionStepKind(graph),
			GraphID: graph.InstanceID,
		})

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
				session.Steps = append(session.Steps, SessionStep{
					Kind:    SessionStepEnterRecoveryCode,
					GraphID: graph.InstanceID,
				})
			case *nodes.EdgeAuthenticationPassword:
				session.Steps = append(session.Steps, SessionStep{
					Kind:    SessionStepEnterPassword,
					GraphID: graph.InstanceID,
				})
			case *nodes.EdgeAuthenticationTOTP:
				session.Steps = append(session.Steps, SessionStep{
					Kind:    SessionStepEnterTOTP,
					GraphID: graph.InstanceID,
				})
			case *nodes.EdgeAuthenticationOOBTrigger:
				inputFn = func() (input interface{}, err error) {
					input = &inputTriggerOOB{AuthenticatorIndex: 0}
					return
				}
			default:
				panic(fmt.Errorf("webapp: unexpected edge: %T", defaultEdge))
			}

		case SessionStepCreateAuthenticator:
			switch defaultEdge := edges[0].(type) {
			case *nodes.EdgeCreateAuthenticatorPassword:
				session.Steps = append(session.Steps, SessionStep{
					Kind:    SessionStepCreatePassword,
					GraphID: graph.InstanceID,
				})
			case *nodes.EdgeCreateAuthenticatorOOBSetup:
				session.Steps = append(session.Steps, SessionStep{
					Kind:    SessionStepSetupOOBOTP,
					GraphID: graph.InstanceID,
				})
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
	result *Result,
	inputFn func() (interface{}, error),
	graphFn func(ctx *interaction.Context) (*interaction.Graph, error),
) (*interaction.Graph, []interaction.Edge, error) {
	var graph *interaction.Graph
	var edges []interaction.Edge
	interactionErr := s.Graph.DryRun("", func(ctx *interaction.Context) (*interaction.Graph, error) {
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
			result.cookies = append(result.cookies, clearCookie.Cookies...)
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
		result.cookies = append(result.cookies, errCookie)
	} else if isFinished {
		// Loop from start to end to collect cookies.
		// This iteration order allows newer node to overwrite cookies.
		for _, node := range graph.Nodes {
			if a, ok := node.(CookiesGetter); ok {
				result.cookies = append(result.cookies, a.GetCookies()...)
			}
		}
	}

	// Transition to redirect URI
	if isFinished {
		result.redirectURI = session.RedirectURI
	} else if interactionErr == nil {
		if a, ok := graph.CurrentNode().(interface{ GetRedirectURI() string }); ok {
			result.redirectURI = a.GetRedirectURI()
		} else {
			result.redirectURI = session.CurrentStep().URL().String()
		}
	} else {
		if a, ok := graph.CurrentNode().(interface{ GetErrorRedirectURI() string }); ok {
			result.redirectURI = a.GetErrorRedirectURI()
		} else {
			u := url.URL{
				Path:     s.Request.URL.Path,
				RawQuery: s.Request.URL.RawQuery,
			}
			result.redirectURI = u.String()
		}
	}
	s.Logger.Debugf("interaction: redirect to" + result.redirectURI)

	// Collect extras
	session.Extra = collectExtras(graph.CurrentNode())

	// Persist/discard session
	if isFinished && !session.KeepAfterFinish {
		err := s.Sessions.Delete(session.ID)
		if err != nil {
			return err
		}
		result.cookies = append(result.cookies, s.CookieFactory.ClearCookie(s.SessionCookie.Def))
	} else if isNewGraph {
		err := s.Sessions.Create(session)
		if err != nil {
			return err
		}
		result.cookies = append(result.cookies, s.CookieFactory.ValueCookie(s.SessionCookie.Def, session.ID))
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
	switch graph.CurrentNode().(type) {
	case *nodes.NodeUseIdentityOAuthProvider:
		return SessionStepOAuthRedirect
	case *nodes.NodeDoUseUser:
		switch graph.Intent.(type) {
		case *intents.IntentRemoveIdentity:
			return SessionStepEnterLoginID
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeUpdateIdentityBegin:
		return SessionStepEnterLoginID
	case *nodes.NodeCreateIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindPromote:
				return SessionStepPromoteUser
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		case *intents.IntentAddIdentity:
			return SessionStepEnterLoginID
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeAuthenticationBegin:
		return SessionStepAuthenticate
	case *nodes.NodeCreateAuthenticatorBegin:
		return SessionStepCreateAuthenticator
	case *nodes.NodeAuthenticationOOBTrigger:
		return SessionStepEnterOOBOTP
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		return SessionStepEnterOOBOTP
	case *nodes.NodeCreateAuthenticatorTOTPSetup:
		return SessionStepSetupTOTP
	case *nodes.NodeGenerateRecoveryCodeBegin:
		return SessionStepSetupRecoveryCode
	case *nodes.NodeVerifyIdentity:
		return SessionStepVerifyIdentity
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
