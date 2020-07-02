package hook

import (
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
)

//go:generate mockgen -source=mutator.go -destination=mutator_mock_test.go -package hook

type Mutator interface {
	Add(event.Mutations) error
	Apply() error
}

type MutatorFactory struct {
	Users UserProvider
}

func (f *MutatorFactory) New(e *event.Event, u *model.User) Mutator {
	return &mutator{
		Users:     f.Users,
		Event:     e,
		User:      u,
		Mutations: event.Mutations{},
	}
}

type mutator struct {
	Users UserProvider

	Event     *event.Event
	User      *model.User
	Mutations event.Mutations
}

func (m *mutator) Add(mutations event.Mutations) error {
	m.Mutations = m.Mutations.WithMutationsApplied(mutations)
	if payload, ok := m.Event.Payload.(event.UserAwarePayload); ok {
		m.Event.Payload = payload.WithMutationsApplied(m.Mutations)
	}
	m.Mutations.ApplyToUser(m.User)
	return nil
}

func (m *mutator) Apply() error {
	mutations := m.Mutations

	// mutate user profile
	if mutations.IsNoop() {
		return nil
	}

	if mutations.Metadata != nil {
		err := m.Users.UpdateMetadata(m.User, *mutations.Metadata)
		if err != nil {
			return err
		}
		mutations.Metadata = nil
	}

	return nil
}
