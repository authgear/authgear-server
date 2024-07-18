package nodes

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityAnonymous{})
}

type InputUseIdentityAnonymous interface {
	GetAnonymousRequestToken() string
	SignUpAnonymousUserWithoutKey() bool
	GetPromotionCode() string
}

type EdgeUseIdentityAnonymous struct {
	IsAuthentication bool
}

// nolint:gocognit
func (e *EdgeUseIdentityAnonymous) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityAnonymous
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	enabled := false
	for _, t := range ctx.Config.Authentication.Identities {
		if t == model.IdentityTypeAnonymous {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, api.NewInvariantViolated(
			"AnonymousUserDisallowed",
			"anonymous users are not allowed",
			nil,
		)
	}

	if input.SignUpAnonymousUserWithoutKey() {
		if !e.IsAuthentication {
			panic("interaction: SignUpAnonymousUserWithoutKey should be used for signup only")
		}

		spec := &identity.Spec{
			Type: model.IdentityTypeAnonymous,
			Anonymous: &identity.AnonymousSpec{
				KeyID: "",
				Key:   "",
			},
		}

		return &NodeUseIdentityAnonymous{
			IsAuthentication: e.IsAuthentication,
			IdentitySpec:     spec,
		}, nil
	}

	promotionCode := input.GetPromotionCode()
	if promotionCode != "" {
		// promote user with promotion code flow
		if e.IsAuthentication {
			panic("interaction: cannot use promotion code for authentication")
		}

		codeObj, err := ctx.AnonymousUserPromotionCodeStore.GetPromotionCode(anonymous.HashPromotionCode(promotionCode))
		if err != nil {
			return nil, err
		}

		promoteUserID := codeObj.UserID
		promoteIdentityID := codeObj.IdentityID

		anonIdentity, err := ctx.AnonymousIdentities.Get(promoteUserID, promoteIdentityID)
		if err != nil {
			panic(fmt.Errorf("interaction: failed to fetch anonymous identity: %s, %s, %w", promoteUserID, promoteIdentityID, err))
		}

		if anonIdentity.KeyID != "" {
			panic(fmt.Errorf("interaction: anonymous user with key should use jwt to trigger promotion flow"))
		}

		spec := &identity.Spec{
			Type: model.IdentityTypeAnonymous,
			Anonymous: &identity.AnonymousSpec{
				ExistingUserID:     anonIdentity.UserID,
				ExistingIdentityID: anonIdentity.ID,
			},
		}

		return &NodeUseIdentityAnonymous{
			PromotionCode:    promotionCode,
			IsAuthentication: e.IsAuthentication,
			IdentitySpec:     spec,
		}, nil
	}

	jwt := input.GetAnonymousRequestToken()
	request, err := ctx.AnonymousIdentities.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, api.ErrInvalidCredentials
	}

	chal, err := ctx.Challenges.Get(request.Challenge)
	if err != nil || chal.Purpose != challenge.PurposeAnonymousRequest {
		return nil, api.ErrInvalidCredentials
	}

	anonIdentity, err := ctx.AnonymousIdentities.GetByKeyID(request.KeyID)
	if errors.Is(err, api.ErrIdentityNotFound) {
		anonIdentity = nil
	} else if err != nil {
		return nil, err
	}

	existingUserID := ""
	existingIdentityID := ""
	if anonIdentity != nil {
		existingUserID = anonIdentity.UserID
		existingIdentityID = anonIdentity.ID
		// Key ID has associated identity =>
		// verify the JWT signature before proceeding to use the key ID.
		request, err = ctx.AnonymousIdentities.ParseRequest(jwt, anonIdentity)
		if err != nil {
			dispatchEvent := func() error {
				userID := anonIdentity.UserID
				userRef := model.UserRef{
					Meta: model.Meta{
						ID: userID,
					},
				}
				err = ctx.Events.DispatchEventOnCommit(&nonblocking.AuthenticationFailedIdentityEventPayload{
					UserRef:      userRef,
					IdentityType: string(model.IdentityTypeAnonymous),
				})
				if err != nil {
					return err
				}

				return nil
			}
			_ = dispatchEvent()
			return nil, api.ErrInvalidCredentials
		}
	} else if request.Key == nil {
		// No associated identity => a new key must be provided.
		return nil, api.ErrInvalidCredentials
	}

	key, err := json.Marshal(request.Key)
	if err != nil {
		return nil, err
	}

	spec := &identity.Spec{
		Type: model.IdentityTypeAnonymous,
		Anonymous: &identity.AnonymousSpec{
			ExistingUserID:     existingUserID,
			ExistingIdentityID: existingIdentityID,
			KeyID:              request.KeyID,
			Key:                string(key),
		},
	}

	return &NodeUseIdentityAnonymous{
		JWT:              jwt,
		IsAuthentication: e.IsAuthentication,
		IdentitySpec:     spec,
	}, nil
}

type NodeUseIdentityAnonymous struct {
	IsAuthentication bool           `json:"is_authentication"`
	IdentitySpec     *identity.Spec `json:"identity_spec"`
	PromotionCode    string         `json:"promotion_code,omitempty"`
	JWT              string         `json:"jwt,omitempty"`
}

func (n *NodeUseIdentityAnonymous) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityAnonymous) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.JWT != "" {
				request, err := ctx.AnonymousIdentities.ParseRequestUnverified(n.JWT)
				if err != nil {
					return err
				}

				_, err = ctx.Challenges.Consume(request.Challenge)
				if err != nil {
					return err
				}
			}

			if n.PromotionCode != "" {
				codeObj, err := ctx.AnonymousUserPromotionCodeStore.GetPromotionCode(anonymous.HashPromotionCode(n.PromotionCode))
				if err != nil {
					return err
				}

				err = ctx.AnonymousUserPromotionCodeStore.DeletePromotionCode(codeObj)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeUseIdentityAnonymous) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec, IsAuthentication: n.IsAuthentication}}, nil
}
