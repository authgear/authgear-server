package nodes

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityBiometric{})
}

type InputUseIdentityBiometric interface {
	GetBiometricRequestToken() string
}

type EdgeUseIdentityBiometric struct {
	IsCreating bool
}

func (e *EdgeUseIdentityBiometric) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityBiometric
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	enabled := false
	for _, t := range ctx.Config.Authentication.Identities {
		if t == authn.IdentityTypeBiometric {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, interaction.NewInvariantViolated(
			"BiometricDisallowed",
			"biometric is not allowed",
			nil,
		)
	}

	jwt := input.GetBiometricRequestToken()

	request, err := ctx.BiometricIdentities.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, interaction.ErrInvalidCredentials
	}

	purpose, err := ctx.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeBiometricRequest {
		return nil, interaction.ErrInvalidCredentials
	}

	var iden *identitybiometric.Identity
	switch request.Action {
	case identitybiometric.RequestActionSetup:
		// FIXME(biometric): validate device info
		if request.Key == nil {
			return nil, interaction.ErrInvalidCredentials
		}
	case identitybiometric.RequestActionAuthenticate:
		iden, err = ctx.BiometricIdentities.GetByKeyID(request.KeyID)
		if err != nil {
			return nil, interaction.ErrInvalidCredentials
		}
		request, err = ctx.BiometricIdentities.ParseRequest(jwt, iden)
		if err != nil {
			return nil, interaction.ErrInvalidCredentials
		}
	}

	key, err := json.Marshal(request.Key)
	if err != nil {
		return nil, err
	}

	spec := &identity.Spec{
		Type: authn.IdentityTypeBiometric,
		Claims: map[string]interface{}{
			identity.IdentityClaimBiometricKeyID:      request.KeyID,
			identity.IdentityClaimBiometricKey:        string(key),
			identity.IdentityClaimBiometricDeviceInfo: request.DeviceInfo,
		},
	}

	return &NodeUseIdentityBiometric{
		IsCreating:   e.IsCreating,
		IdentitySpec: spec,
	}, nil
}

type NodeUseIdentityBiometric struct {
	IsCreating   bool           `json:"is_creating"`
	IdentitySpec *identity.Spec `json:"identity_spec"`
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
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
}
