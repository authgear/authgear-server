package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

var ChangePasswordFailed = apierrors.Invalid.WithReason("ChangePasswordFailed")
var ErrNoPassword = ChangePasswordFailed.NewWithCause("the user does not have a password", apierrors.StringCause("NoPassword"))

func init() {
	interaction.RegisterNode(&NodeChangePasswordBegin{})
	interaction.RegisterNode(&NodeChangePasswordEnd{})
}

type EdgeChangePasswordBegin struct {
	Stage interaction.AuthenticationStage `json:"stage"`
}

func (e *EdgeChangePasswordBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeChangePasswordBegin{
		Stage: e.Stage,
	}, nil
}

type NodeChangePasswordBegin struct {
	Stage interaction.AuthenticationStage `json:"stage"`
}

func (n *NodeChangePasswordBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeChangePasswordBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeChangePasswordBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeChangePassword{
		Stage: n.Stage,
	}}, nil
}

type InputChangePassword interface {
	GetOldPassword() string
	GetNewPassword() string
}

type EdgeChangePassword struct {
	Stage interaction.AuthenticationStage
}

func (e *EdgeChangePassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (node interaction.Node, err error) {
	var input InputChangePassword
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	oldPassword := input.GetOldPassword()
	newPassword := input.GetNewPassword()

	userID := graph.MustGetUserID()
	ais, err := ctx.Authenticators.List(
		userID,
		authenticator.KeepType(authn.AuthenticatorTypePassword),
		authenticator.KeepKind(stageToAuthenticatorKind(e.Stage)),
	)
	if err != nil {
		return
	}

	if len(ais) == 0 {
		err = ErrNoPassword
		return
	}

	if len(ais) != 1 {
		err = fmt.Errorf("changepassword: detected user %s having more than 1 password", userID)
		return
	}
	oldInfo := ais[0]

	err = ctx.Authenticators.VerifySecret(oldInfo, nil, oldPassword)
	if err != nil {
		err = interaction.ErrInvalidCredentials
		return
	}

	changed, newInfo, err := ctx.Authenticators.WithSecret(oldInfo, newPassword)
	if err != nil {
		return
	}

	newNode := &NodeChangePasswordEnd{
		Stage:            e.Stage,
		OldAuthenticator: oldInfo,
	}
	if changed {
		newNode.NewAuthenticator = newInfo
	}

	node = newNode
	return node, nil
}

type NodeChangePasswordEnd struct {
	Stage            interaction.AuthenticationStage `json:"stage"`
	OldAuthenticator *authenticator.Info             `json:"old_authenticator"`
	NewAuthenticator *authenticator.Info             `json:"new_authenticator,omitempty"`
}

func (n *NodeChangePasswordEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeChangePasswordEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeChangePasswordEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	if n.NewAuthenticator != nil {
		return []interaction.Edge{
			&EdgeDoUpdateAuthenticator{
				Stage:                     n.Stage,
				AuthenticatorBeforeUpdate: n.OldAuthenticator,
				AuthenticatorAfterUpdate:  n.NewAuthenticator,
			},
		}, nil
	}

	return graph.Intent.DeriveEdgesForNode(graph, n)
}
