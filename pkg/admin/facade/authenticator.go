package facade

import (
	"context"
	"sort"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
)

type AuthenticatorService interface {
	New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, ai *authenticator.Info, options *service.UpdatePasswordOptions) (bool, *authenticator.Info, error)

	Create(ctx context.Context, info *authenticator.Info) error
	Update(ctx context.Context, info *authenticator.Info) error
	Get(ctx context.Context, id string) (*authenticator.Info, error)
	Count(ctx context.Context, userID string) (uint64, error)
	ListRefsByUsers(ctx context.Context, userIDs []string, authenticatorType *apimodel.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error)
}

type AuthenticatorFacade struct {
	Authenticators AuthenticatorService
	Interaction    InteractionService
}

func (f *AuthenticatorFacade) Get(ctx context.Context, id string) (*authenticator.Info, error) {
	return f.Authenticators.Get(ctx, id)
}

func (f *AuthenticatorFacade) List(ctx context.Context, userID string, authenticatorType *apimodel.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error) {
	refs, err := f.Authenticators.ListRefsByUsers(ctx, []string{userID}, authenticatorType, authenticatorKind)
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

func (f *AuthenticatorFacade) Remove(ctx context.Context, authenticatorInfo *authenticator.Info) error {
	_, err := f.Interaction.Perform(
		ctx,
		interactionintents.NewIntentRemoveAuthenticator(authenticatorInfo.UserID),
		&removeAuthenticatorInput{authenticatorInfo: authenticatorInfo},
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *AuthenticatorFacade) CreateBySpec(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error) {
	info, err := f.Authenticators.New(ctx, spec)
	if err != nil {
		return nil, err
	}
	err = f.Authenticators.Create(ctx, info)

	if err != nil {
		return nil, err
	}
	return info, err
}
