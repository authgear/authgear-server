package flows

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/log"
)

//go:generate mockgen -source=state_service.go -destination=state_service_mock_test.go -package flows

type StateStore interface {
	Get(id string) (*State, error)
	Set(state *State) error
	Delete(id string) error
}

type StateServiceLogger struct{ *log.Logger }

func NewStateServiceLogger(lf *log.Factory) StateServiceLogger {
	return StateServiceLogger{lf.New("interactionflows-state")}
}

type StateService struct {
	ServerConfig *config.ServerConfig
	StateStore   StateStore
	Logger       StateServiceLogger
}

func (p *StateService) MakeState(r *http.Request) *State {
	s := NewState()
	r.Form.Set("x_sid", s.ID)
	q := r.URL.Query()
	q.Set("x_sid", s.ID)
	r.URL.RawQuery = q.Encode()
	return s
}

func (p *StateService) CreateState(s *State, redirectURI string) *State {
	s.Extra[ExtraRedirectURI] = redirectURI
	err := p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}
	return s
}

func (p *StateService) UpdateState(s *State, r *WebAppResult, inputError error) {
	if s == nil {
		panic("webapp: expected non-nil state to update")
	}

	s.SetError(inputError)
	if inputError != nil && !skyerr.IsAPIError(inputError) {
		p.Logger.WithError(inputError).Error("unexpected error occurred")
	}

	err := p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}
}

func (p *StateService) RestoreState(r *http.Request, optional bool) (state *State, err error) {
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

func (p *StateService) DeleteState(s *State) {
	err := p.StateStore.Delete(s.ID)
	if err != nil {
		p.Logger.WithError(err).Error("failed to delete web app state")
	}
}
