package facade

import (
	"errors"
	"sort"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type IdentityService interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
	ListRefsByUsers(userIDs []string) ([]*identity.Ref, error)
}

type IdentityFacade struct {
	Identities  IdentityService
	Interaction InteractionService
}

func (f *IdentityFacade) Get(ref *identity.Ref) (*identity.Info, error) {
	return f.Identities.Get(ref.UserID, ref.Type, ref.ID)
}

func (f *IdentityFacade) List(userID string) ([]*identity.Ref, error) {
	refs, err := f.Identities.ListRefsByUsers([]string{userID})
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

func (f *IdentityFacade) Create(userID string, identityDef model.IdentityDef, password string) (*identity.Info, error) {
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
			// TODO(interaction): better interpretation of input required error?
			return nil, interaction.NewInvariantViolated(
				"PasswordRequired",
				"password is required",
				nil,
			)
		}
	}
	if err != nil {
		return nil, err
	}

	return graph.GetUserNewIdentities()[0], nil
}
