package facade

import (
	"sort"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
)

type AuthenticatorService interface {
	New(spec *authenticator.Spec) (*authenticator.Info, error)
	Create(info *authenticator.Info) error
	Get(id string) (*authenticator.Info, error)
	Count(userID string) (uint64, error)
	ListRefsByUsers(userIDs []string, authenticatorType *apimodel.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error)
}

type AuthenticatorFacade struct {
	Authenticators AuthenticatorService
	Interaction    InteractionService
}

func (f *AuthenticatorFacade) Get(id string) (*authenticator.Info, error) {
	return f.Authenticators.Get(id)
}

func (f *AuthenticatorFacade) List(userID string, authenticatorType *apimodel.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error) {
	refs, err := f.Authenticators.ListRefsByUsers([]string{userID}, authenticatorType, authenticatorKind)
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

func (f *AuthenticatorFacade) CreateBySpec(spec *authenticator.Spec) (*authenticator.Info, error) {
	info, err := f.Authenticators.New(spec)
	if err != nil {
		return nil, err
	}
	err = f.Authenticators.Create(info)

	if err != nil {
		return nil, err
	}
	return info, err
}
