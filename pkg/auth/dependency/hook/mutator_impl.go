package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type LoginIDProvider interface {
	List(userID string) ([]*loginid.Identity, error)
}

type mutatorImpl struct {
	Event             *event.Event
	User              *model.User
	LoginIDIdentities *[]*loginid.Identity
	Mutations         event.Mutations
	Time              time.Provider

	UserVerificationConfig *config.UserVerificationConfiguration
	LoginIDs               LoginIDProvider
	Users                  UserProvider
}

func NewMutator(
	verifyConfig *config.UserVerificationConfiguration,
	loginIDProvider LoginIDProvider,
	up UserProvider,
) Mutator {
	return &mutatorImpl{
		UserVerificationConfig: verifyConfig,
		LoginIDs:               loginIDProvider,
		Users:                  up,
	}
}

func (mutator *mutatorImpl) New(ev *event.Event, user *model.User) Mutator {
	newMutator := *mutator
	newMutator.Event = ev
	newMutator.User = user
	newMutator.Mutations = event.Mutations{}
	return &newMutator
}

func (mutator *mutatorImpl) Add(mutations event.Mutations) error {
	mutator.Mutations = mutator.Mutations.WithMutationsApplied(mutations)
	if payload, ok := mutator.Event.Payload.(event.UserAwarePayload); ok {
		mutator.Event.Payload = payload.WithMutationsApplied(mutator.Mutations)
	}
	mutator.Mutations.ApplyToUser(mutator.User)
	return nil
}

func (mutator *mutatorImpl) Apply() error {
	mutations := mutator.Mutations

	// mutate user profile
	if mutations.IsNoop() {
		return nil
	}

	if mutations.Metadata != nil {
		err := mutator.Users.UpdateMetadata(mutator.User, *mutations.Metadata)
		if err != nil {
			return err
		}
		mutations.Metadata = nil
	}

	return nil
}
