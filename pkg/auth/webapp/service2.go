package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type SessionStore interface {
	Create(session *Session) (err error)
	Update(session *Session) (err error)
}

type Service2 struct {
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

func (s *Service2) GetWithIntent(path string, session *Session, intent interaction.Intent) (*interaction.Graph, error) {
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

	session.Steps = append(session.Steps, SessionStep{
		GraphID: graph.InstanceID,
		Path:    path,
	})
	return graph, nil
}

func (s *Service2) PostWithIntent(
	path string,
	session *Session,
	intent interaction.Intent,
	inputFn func() (interface{}, error),
) (result *Result, err error) {
	panic("not implemented")
}

func (s *Service2) PostWithInput(
	path string,
	session *Session,
	inputFn func() (interface{}, error),
) (result *Result, err error) {
	panic("not implemented")
}
