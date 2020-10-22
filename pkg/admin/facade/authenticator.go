package facade

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
)

type AuthenticatorService interface {
	Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error)
	Count(userID string) (uint64, error)
	ListRefsByUsers(userIDs []string) ([]*authenticator.Ref, error)
}

type AuthenticatorFacade struct {
	Authenticators AuthenticatorService
	Interaction    InteractionService
}

func (f *AuthenticatorFacade) Get(ref *authenticator.Ref) (*authenticator.Info, error) {
	return f.Authenticators.Get(ref.UserID, ref.Type, ref.ID)
}

func (f *AuthenticatorFacade) List(userID string) ([]*authenticator.Ref, error) {
	refs, err := f.Authenticators.ListRefsByUsers([]string{userID})
	if err != nil {
		return nil, err
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].CreatedAt != refs[j].CreatedAt {
			return refs[i].CreatedAt.Before(refs[j].CreatedAt)
		}
		return refs[i].ID < refs[j].ID
	})

	return refs, nil
}

func (f *AuthenticatorFacade) Remove(authenticatorInfo *authenticator.Info) error {
	_, err := f.Interaction.Perform(
		interactionintents.NewIntentRemoveAuthenticator(authenticatorInfo.UserID),
		&removeAuthenticatorInput{authenticatorInfo: authenticatorInfo},
	)
	if err != nil {
		return err
	}

	return nil
}
