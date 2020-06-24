package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/log"
)

//go:generate mockgen -source=state_provider.go -destination=state_provider_mock_test.go -package webapp

type StateStore interface {
	Get(id string) (*State, error)
	Set(state *State) error
}

type StateProvider interface {
	CreateState(r *http.Request, inputError error)
	UpdateState(r *http.Request, inputError error)
	UpdateError(id string, inputError error)
	RestoreState(r *http.Request, optional bool) (state *State, err error)
}

type StateProviderLogger struct{ *log.Logger }

func NewStateProviderLogger(lf *log.Factory) StateProviderLogger {
	return StateProviderLogger{lf.New("webapp-state")}
}

type StateProviderImpl struct {
	StateStore StateStore
	Logger     StateProviderLogger
}

func (p *StateProviderImpl) UpdateError(id string, inputError error) {
	s, err := p.StateStore.Get(id)
	if err != nil {
		panic(err)
	}
	s.SetError(inputError)
	if inputError != nil && !skyerr.IsAPIError(inputError) {
		p.Logger.WithError(inputError).Error("unexpected error occurred")
	}

	err = p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}
}

func (p *StateProviderImpl) CreateState(r *http.Request, inputError error) {
	s := NewState()

	r.Form.Set("x_sid", s.ID)
	s.SetForm(r.Form)
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
}

func (p *StateProviderImpl) UpdateState(r *http.Request, inputError error) {
	sid := r.URL.Query().Get("x_sid")

	s, err := p.StateStore.Get(sid)
	if err != nil {
		panic(err)
	}

	s.SetForm(r.Form)
	s.SetError(inputError)
	if inputError != nil && !skyerr.IsAPIError(inputError) {
		p.Logger.WithError(inputError).Error("unexpected error occurred")
	}

	err = p.StateStore.Set(s)
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

	change, err := state.ClearErrorIfFormChanges(r.Form)
	if err != nil {
		return
	}

	if change {
		err = p.StateStore.Set(state)
		if err != nil {
			return
		}
	}

	err = state.Restore(r.Form)
	if err != nil {
		return
	}

	return
}
