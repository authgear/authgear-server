package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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
	Logger        ServiceLogger
	Request       *http.Request
	Sessions      SessionStore
	SessionCookie SessionCookieDef
	ErrorCookie   *ErrorCookie
	CookieFactory CookieFactory

	Graph GraphService
}

func (s *Service2) rewindSessionHistory(session *Session, path string) *SessionStep {
	step, err := strconv.Atoi(s.Request.URL.Query().Get("x_step"))
	if err != nil {
		step = 0
	}
	if step < 0 {
		step = 0
	} else if step > len(session.Steps) {
		step = len(session.Steps)
	}
	session.Steps = session.Steps[:step]

	if len(session.Steps) == 0 {
		return nil
	}

	curStep := session.CurrentStep()
	if curStep.Path != path {
		return nil
	}
	return &curStep
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
	var graph *interaction.Graph
	var edges []interaction.Edge
	var clearCookie *interaction.ErrClearCookie
	err = s.Graph.DryRun(session.ID, func(ctx *interaction.Context) (newGraph *interaction.Graph, err error) {
		graph, err = s.Graph.NewGraph(ctx, intent)
		if err != nil {
			return
		}

		input, err := inputFn()
		if err != nil {
			return
		}

		newGraph, edges, err = s.Graph.Accept(ctx, graph, input)

		errors.As(err, &clearCookie)

		var inputRequired *interaction.ErrInputRequired
		if errors.As(err, &inputRequired) {
			err = nil
			graph = newGraph
			return
		}
		if err != nil {
			return
		}

		graph = newGraph
		return
	})

	return s.afterPost(session, graph, edges, err, clearCookie, true)
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

	var edges []interaction.Edge
	graph, err := s.Graph.Get(step.GraphID)
	if err != nil {
		return nil, err
	}

	var clearCookie *interaction.ErrClearCookie
	err = s.Graph.DryRun(session.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
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

		errors.As(err, &clearCookie)

		var inputRequired *interaction.ErrInputRequired
		if errors.As(err, &inputRequired) {
			graph = newGraph
			return newGraph, nil
		} else if err != nil {
			return nil, err
		}

		graph = newGraph
		return newGraph, nil
	})

	return s.afterPost(session, graph, edges, err, clearCookie, false)
}

func (s *Service2) afterPost(
	session *Session,
	graph *interaction.Graph,
	edges []interaction.Edge,
	interactionErr error,
	clearCookie *interaction.ErrClearCookie,
	isNewIntent bool,
) (*Result, error) {
	finished := len(edges) == 0 && interactionErr == nil
	// The graph finished. Apply its effect permanently
	if finished {
		s.Logger.Debugf("afterPost run graph")
		interactionErr = s.Graph.Run(session.ID, graph)
	}

	// Populate cookies.
	var cookies []*http.Cookie
	if clearCookie != nil {
		cookies = append(cookies, clearCookie.Cookies...)
	}
	if interactionErr != nil {
		if !apierrors.IsAPIError(interactionErr) {
			s.Logger.Errorf("afterPost error: %v", interactionErr)
		}
		errCookie, err := s.ErrorCookie.SetError(apierrors.AsAPIError(interactionErr))
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, errCookie)
	}
	if finished {
		// Loop from start to end to collect cookies.
		// This iteration order allows newer node to overwrite cookies.
		for _, node := range graph.Nodes {
			if a, ok := node.(CookiesGetter); ok {
				cookies = append(cookies, a.GetCookies()...)
			}
		}
	}

	// Transition to redirect URI.
	redirectURI, isExternalRedirect := deriveRedirectURI(interactionErr, graph)
	if !isExternalRedirect {
		// Internal redirect: redirectURI is a path.
		if interactionErr == nil {
			// Push step into history if success
			session.Steps = append(session.Steps, SessionStep{
				GraphID: graph.InstanceID,
				Path:    redirectURI,
			})
		}

		if len(session.Steps) > 0 {
			// Append current step index to redirect URI
			query := url.Values{}
			query.Set("x_step", strconv.Itoa(len(session.Steps)))
			uri := url.URL{
				Path:     redirectURI,
				RawQuery: query.Encode(),
			}
			redirectURI = uri.String()
		}
	}

	// Persist/discard session
	if finished && !session.KeepAfterFinish {
		err := s.Sessions.Delete(session.ID)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, s.CookieFactory.ClearCookie(s.SessionCookie.Def))
	} else if isNewIntent {
		err := s.Sessions.Create(session)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, s.CookieFactory.ValueCookie(s.SessionCookie.Def, session.ID))
	} else {
		err := s.Sessions.Update(session)
		if err != nil {
			return nil, err
		}
	}

	return &Result{
		redirectURI: redirectURI,
		cookies:     cookies,
	}, nil
}

// nolint:gocyclo
func deriveRedirectURI(err error, graph *interaction.Graph) (uri string, isExternal bool) {
	node := graph.CurrentNode()
	if err != nil {
		if a, ok := node.(interface{ GetErrorRedirectURI() string }); ok {
			return a.GetErrorRedirectURI(), true
		}
	} else {
		if a, ok := node.(interface{ GetRedirectURI() string }); ok {
			return a.GetRedirectURI(), true
		}
	}

	var path string
	switch node.(type) {
	case *nodes.NodeSelectIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindLogin:
				path = "/login"
			case intents.IntentAuthenticateKindSignup:
				path = "/signup"
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeDoUseUser:
		switch graph.Intent.(type) {
		case *intents.IntentRemoveIdentity:
			path = "/enter_login_id"
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeUpdateIdentityBegin:
		path = "/enter_login_id"
	case *nodes.NodeCreateIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindPromote:
				path = "/promote_user"
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		case *intents.IntentAddIdentity:
			path = "/enter_login_id"
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeAuthenticationBegin:
		// Ideally we should use 307 but if we use 307
		// the CSRF middleware will complain about invalid CSRF token.
		path = "/authentication_begin"
	case *nodes.NodeCreateAuthenticatorBegin:
		path = "/create_authenticator_begin"
	case *nodes.NodeAuthenticationOOBTrigger:
		path = "/enter_oob_otp"
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		path = "/enter_oob_otp"
	case *nodes.NodeCreateAuthenticatorTOTPSetup:
		path = "/setup_totp"
	case *nodes.NodeGenerateRecoveryCodeBegin:
		path = "/setup_recovery_code"
	case *nodes.NodeVerifyIdentity:
		path = "/verify_identity"
	default:
		panic(fmt.Errorf("webapp: unexpected node: %T", graph.CurrentNode()))
	}

	return path, false
}
