package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/log"
)

//go:generate mockgen -source=state_provider.go -destination=state_provider_mock_test.go -package webapp

type StateStore interface {
	Get(id string) (*State, error)
	Set(state *State) error
	Delete(id string) error
}

type StateProvider interface {
	CreateState(r *http.Request, result *interactionflows.WebAppResult, inputError error) *State
	UpdateState(s *State, r *interactionflows.WebAppResult, inputError error)
	RestoreState(r *http.Request, optional bool) (state *State, err error)
	DeleteState(s *State)
}

type StateProviderLogger struct{ *log.Logger }

func NewStateProviderLogger(lf *log.Factory) StateProviderLogger {
	return StateProviderLogger{lf.New("webapp-state")}
}

type StateProviderImpl struct {
	StateStore StateStore
	Logger     StateProviderLogger
}

func (p *StateProviderImpl) CreateState(r *http.Request, result *interactionflows.WebAppResult, inputError error) *State {
	s := NewState()

	r.Form.Set("x_sid", s.ID)
	if result != nil {
		s.Interaction = result.Interaction
	}
	s.SetError(inputError)
	if inputError != nil && !skyerr.IsAPIError(inputError) {
		p.Logger.WithError(inputError).Error("unexpected error occurred")
	}

	err := p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}

	q := r.URL.Query()
	q.Set("x_sid", s.ID)
	r.URL.RawQuery = q.Encode()

	return s
}

func (p *StateProviderImpl) UpdateState(s *State, r *interactionflows.WebAppResult, inputError error) {
	if s == nil {
		panic("webapp: expected non-nil state to update")
	}

	var i *interaction.Interaction
	if r != nil && r.Interaction != nil {
		i = r.Interaction
	} else if s.Interaction != nil {
		i = s.Interaction
	}
	s.Interaction = i

	s.SetError(inputError)
	if inputError != nil && !skyerr.IsAPIError(inputError) {
		p.Logger.WithError(inputError).Error("unexpected error occurred")
	}

	err := p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}
}

func (p *StateProviderImpl) RestoreState(r *http.Request, optional bool) (state *State, err error) {
	sid := r.URL.Query().Get("x_sid")

	if optional && sid == "" {
		return
	}

	state, err = p.StateStore.Get(sid)
	if err != nil {
		return
	}

	return
}

func (p *StateProviderImpl) DeleteState(s *State) {
	err := p.StateStore.Delete(s.ID)
	if err != nil {
		p.Logger.WithError(err).Error("failed to delete web app state")
	}
}
