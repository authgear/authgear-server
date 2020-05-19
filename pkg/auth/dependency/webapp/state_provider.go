package webapp

import (
	"net/http"
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

type StateProviderImpl struct {
	StateStore StateStore
}

func (p *StateProviderImpl) UpdateError(id string, inputError error) {
	s, err := p.StateStore.Get(id)
	if err != nil {
		panic(err)
	}
	s.SetError(inputError)

	err = p.StateStore.Set(s)
	if err != nil {
		panic(err)
	}
}

func (p *StateProviderImpl) CreateState(r *http.Request, inputError error) {
	s := NewState()

	s.SetForm(r.Form)
	s.SetError(inputError)

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

	err = state.Restore(r.Form)
	if err != nil {
		return
	}

	return
}
