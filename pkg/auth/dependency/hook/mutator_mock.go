package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type mockMutator struct {
	Event         *event.Event
	User          *model.User
	MutationsList []event.Mutations
	IsApplied     bool
	AddError      error
	ApplyError    error
}

func newMockMutator() *mockMutator {
	return &mockMutator{}
}

func (mutator *mockMutator) New(event *event.Event, user *model.User) Mutator {
	// preserve mock error
	addError := mutator.AddError
	applyError := mutator.ApplyError
	*mutator = mockMutator{}
	mutator.AddError = addError
	mutator.ApplyError = applyError

	// return self for testing
	mutator.Event = event
	mutator.User = user
	return mutator
}

func (mutator *mockMutator) Add(mutations event.Mutations) error {
	mutator.MutationsList = append(mutator.MutationsList, mutations)
	return mutator.AddError
}

func (mutator *mockMutator) Apply() error {
	mutator.IsApplied = true
	return mutator.ApplyError
}
