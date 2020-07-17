package flows

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	corerand "github.com/authgear/authgear-server/pkg/core/rand"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/log"
)

//go:generate mockgen -source=state_service.go -destination=state_service_mock_test.go -package flows

type StateStore interface {
	CreateState(state *State) error
	UpdateState(state *State) error
	DeleteState(flowID string) error
	GetState(instanceID string) (*State, error)
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

func (p *StateService) CreateState(s *State, redirectURI string) *State {
	s.Extra[ExtraRedirectURI] = redirectURI
	err := p.StateStore.CreateState(s)
	if err != nil {
		panic(err)
	}
	return s
}

func (p *StateService) UpdateState(s *State, r *WebAppResult, inputError error) {
	if s == nil {
		panic("interaction_flow_webapp: expected non-nil state to update")
	}

	if s.readOnly {
		panic("interaction_flow_webapp: update read-only state")
	}

	s.Error = skyerr.AsAPIError(inputError)
	if inputError != nil && !skyerr.IsAPIError(inputError) {
		p.Logger.WithError(inputError).Error("unexpected error occurred")
	}

	err := p.StateStore.UpdateState(s)
	if err != nil {
		panic(err)
	}
}

func (p *StateService) RestoreReadOnlyState(r *http.Request, optional bool) (state *State, err error) {
	sid := r.URL.Query().Get("x_sid")

	if optional && sid == "" {
		return
	}

	state, err = p.StateStore.GetState(sid)
	if err != nil {
		return
	}

	state.readOnly = true
	return
}

func (p *StateService) CloneState(r *http.Request) (state *State, err error) {
	sid := r.URL.Query().Get("x_sid")

	state, err = p.StateStore.GetState(sid)
	if err != nil {
		return
	}

	state.Error = nil
	instanceID := corerand.StringWithAlphabet(stateIDLength, stateIDAlphabet, corerand.SecureRand)
	state.InstanceID = instanceID
	state.readOnly = false
	return
}

func (p *StateService) DeleteState(s *State) {
	err := p.StateStore.DeleteState(s.FlowID)
	if err != nil {
		p.Logger.WithError(err).Error("failed to delete web app state")
	}
}
