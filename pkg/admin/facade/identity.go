package facade

import (
	"errors"
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/admin/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	ListRefsByUsers(userIDs []string, identityType *apimodel.IdentityType) ([]*apimodel.IdentityRef, error)
}

type IdentityFacade struct {
	Identities  IdentityService
	Interaction InteractionService
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
	var input interface{} = &addIdentityInput{identityDef: identityDef}
	if password != "" {
		input = &addPasswordInput{inner: input, password: password}
	}

	graph, err := f.Interaction.Perform(
		interactionintents.NewIntentAddIdentity(userID),
		input,
	)
	var errInputRequired *interaction.ErrInputRequired
	if errors.As(err, &errInputRequired) {
		switch graph.CurrentNode().(type) {
		case *nodes.NodeCreateAuthenticatorBegin:
			// When we revamp the creation of identity, we will allow
			// creating identity without password.
			// The current implementation of portal knows when to require
			// password, so this error should not happen.
			// When this really happens, the portal has programming error.
			return nil, fmt.Errorf("password is required to create identity")
		}
	}
	if err != nil {
		return nil, err
	}

	return graph.GetUserNewIdentities()[0].ToRef(), nil
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
