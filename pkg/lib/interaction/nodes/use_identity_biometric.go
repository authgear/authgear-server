package nodes

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityBiometric{})
}

type InputUseIdentityBiometric interface {
	GetBiometricRequestToken() string
}

type EdgeUseIdentityBiometric struct {
	IsAuthentication bool
	IsCreating       bool
}

func (e *EdgeUseIdentityBiometric) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityBiometric
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	enabled := false
	for _, t := range ctx.Config.Authentication.Identities {
		if t == model.IdentityTypeBiometric {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, api.ErrBiometricDisallowed
	}

	jwt := input.GetBiometricRequestToken()

	request, err := ctx.BiometricIdentities.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, api.ErrInvalidCredentials
	}

	purpose, err := ctx.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeBiometricRequest {
		return nil, api.ErrInvalidCredentials
	}

	var iden *identity.Biometric
	switch request.Action {
	case identitybiometric.RequestActionSetup:
		displayName := deviceinfo.DeviceModel(request.DeviceInfo)
		if displayName == "" {
			return nil, api.ErrInvalidCredentials
		}
		if request.Key == nil {
			return nil, api.ErrInvalidCredentials
		}
	case identitybiometric.RequestActionAuthenticate:
		iden, err = ctx.BiometricIdentities.GetByKeyID(request.KeyID)
		if err != nil {
			return nil, api.ErrInvalidCredentials
		}
		request, err = ctx.BiometricIdentities.ParseRequest(jwt, iden)
		if err != nil {
			dispatchEvent := func() error {
				userID := iden.UserID
				userRef := model.UserRef{
					Meta: model.Meta{
						ID: userID,
					},
				}
				err = ctx.Events.DispatchEventOnCommit(&nonblocking.AuthenticationFailedIdentityEventPayload{
					UserRef:      userRef,
					IdentityType: string(model.IdentityTypeBiometric),
				})
				if err != nil {
					return err
				}

				return nil
			}
			_ = dispatchEvent()

			return nil, api.ErrInvalidCredentials
		}
	}

	key, err := json.Marshal(request.Key)
	if err != nil {
		return nil, err
	}

	spec := &identity.Spec{
		Type: model.IdentityTypeBiometric,
		Biometric: &identity.BiometricSpec{
			KeyID:      request.KeyID,
			Key:        string(key),
			DeviceInfo: request.DeviceInfo,
		},
	}

	return &NodeUseIdentityBiometric{
		IsAuthentication: e.IsAuthentication,
		IsCreating:       e.IsCreating,
		IdentitySpec:     spec,
	}, nil
}

type NodeUseIdentityBiometric struct {
	IsAuthentication bool           `json:"is_authentication"`
	IsCreating       bool           `json:"is_creating"`
	IdentitySpec     *identity.Spec `json:"identity_spec"`
}

func (n *NodeUseIdentityBiometric) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityBiometric) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentityBiometric) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	if n.IsCreating {
		return []interaction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	}
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec, IsAuthentication: n.IsAuthentication}}, nil
}
