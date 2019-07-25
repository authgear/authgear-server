package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type MockMutator struct {
	Event         *event.Event
	User          *model.User
	MutationsList []event.Mutations
	IsApplied     bool
	AddError      error
	ApplyError    error
}

func NewMockMutator() *MockMutator {
	return &MockMutator{}
}

func (mutator *MockMutator) Reset() {
	*mutator = MockMutator{}
}

func (mutator *MockMutator) New(event *event.Event, user *model.User) Mutator {
	// preserve mock error
	addError := mutator.AddError
	applyError := mutator.ApplyError
	mutator.Reset()
	mutator.AddError = addError
	mutator.ApplyError = applyError

	// return self for testing
	mutator.Event = event
	mutator.User = user
	return mutator
}

func (mutator *MockMutator) Add(mutations event.Mutations) error {
	mutator.MutationsList = append(mutator.MutationsList, mutations)
	return mutator.AddError
}

func (mutator *MockMutator) Apply() error {
	mutator.IsApplied = true
	return mutator.ApplyError
}
