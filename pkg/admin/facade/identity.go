package facade

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
)

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	ListRefsByUsers(userIDs []string, identityType *apimodel.IdentityType) ([]*apimodel.IdentityRef, error)
	CreateByAdmin(userID string, spec *identity.Spec, password string) (*identity.Info, error)
}

type IdentityFacade struct {
	LoginIDConfig *config.LoginIDConfig
	Identities    IdentityService
	Interaction   InteractionService
}

func (f *IdentityFacade) Get(id string) (*identity.Info, error) {
	return f.Identities.Get(id)
}

func (f *IdentityFacade) List(userID string, identityType *apimodel.IdentityType) ([]*apimodel.IdentityRef, error) {
	refs, err := f.Identities.ListRefsByUsers([]string{userID}, identityType)
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

func (f *IdentityFacade) Remove(identityInfo *identity.Info) error {
	_, err := f.Interaction.Perform(
		interactionintents.NewIntentRemoveIdentity(identityInfo.UserID),
		&removeIdentityInput{identityInfo: identityInfo},
	)
	if err != nil {
		return err
	}
	return nil
}

func (f *IdentityFacade) Create(userID string, identityDef model.IdentityDef, password string) (*apimodel.IdentityRef, error) {
	// NOTE: identityDef is assumed to be a login ID since portal only supports login ID
	loginIDInput := identityDef.(*model.IdentityDefLoginID)
	loginIDKeyCofig, ok := f.LoginIDConfig.GetKeyConfig(loginIDInput.Key)
	if !ok {
		return nil, api.NewInvariantViolated("InvalidLoginIDKey", "invalid login ID key", nil)
	}

	identitySpec := &identity.Spec{
		Type: identityDef.Type(),
		LoginID: &identity.LoginIDSpec{
			Key:   loginIDInput.Key,
			Type:  loginIDKeyCofig.Type,
			Value: loginIDInput.Value,
		},
	}

	iden, err := f.Identities.CreateByAdmin(
		userID,
		identitySpec,
		password,
	)
	if err != nil {
		return nil, err
	}

	return iden.ToRef(), nil
}

func (f *IdentityFacade) Update(identityID string, userID string, identityDef model.IdentityDef) (*apimodel.IdentityRef, error) {
	var input interface{} = &updateIdentityInput{identityDef: identityDef}

	_, err := f.Interaction.Perform(
		interactionintents.NewIntentUpdateIdentity(userID, identityID),
		input,
	)

	if err != nil {
		return nil, err
	}

	identity, err := f.Identities.Get(identityID)

	if err != nil {
		return nil, err
	}

	return identity.ToRef(), nil
}
