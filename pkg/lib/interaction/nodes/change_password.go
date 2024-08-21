package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeChangePasswordBegin{})
	interaction.RegisterNode(&NodeChangePasswordEnd{})
}

type EdgeChangePasswordBegin struct {
	Force bool
	Stage authn.AuthenticationStage
}

func (e *EdgeChangePasswordBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeChangePasswordBegin{
		Force: e.Force,
		Stage: e.Stage,
	}, nil
}

type NodeChangePasswordBegin struct {
	Force  bool                                   `json:"force"`
	Reason *interaction.AuthenticatorUpdateReason `json:"reason,omitempty"`
	Stage  authn.AuthenticationStage              `json:"stage"`
}

func (n *NodeChangePasswordBegin) GetChangeReason() *interaction.AuthenticatorUpdateReason {
	return n.Reason
}

func (n *NodeChangePasswordBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeChangePasswordBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeChangePasswordBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeChangePassword{
		Stage:  n.Stage,
		Reason: n.Reason,
	}}, nil
}

func (n *NodeChangePasswordBegin) IsForceChangePassword() bool {
	return n.Force
}

type InputChangePassword interface {
	GetAuthenticationStage() authn.AuthenticationStage
	GetOldPassword() string
	GetNewPassword() string
}

type EdgeChangePassword struct {
	Stage  authn.AuthenticationStage
	Reason *interaction.AuthenticatorUpdateReason
}

func (e *EdgeChangePassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (node interaction.Node, err error) {
	var input InputChangePassword
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	// We have to check the state of the input to ensure
	// the input for this edge.
	// We do not do this, the primary password input will be feeded to
	// the secondary edge.
	// Two passwords will be changed to the same value.
	stage := input.GetAuthenticationStage()
	if stage != e.Stage {
		return nil, interaction.ErrIncompatibleInput
	}

	oldPassword := input.GetOldPassword()
	newPassword := input.GetNewPassword()

	userID := graph.MustGetUserID()
	ais, err := ctx.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(stageToAuthenticatorKind(e.Stage)),
	)
	if err != nil {
		return
	}

	if len(ais) == 0 {
		err = api.ErrNoPassword
		return
	}

	if len(ais) != 1 {
		err = fmt.Errorf("changepassword: detected user %s having more than 1 password", userID)
		return
	}
	oldInfo := ais[0]

	if verifiedAuthenticator, ok := graph.GetUserAuthenticator(e.Stage); ok && verifiedAuthenticator.ID == oldInfo.ID {
		// The password authenticator we are changing has been verified in this interaction.
		// We avoid asking the user to provide the password again.
	} else {
		_, err = ctx.Authenticators.VerifyWithSpec(oldInfo, &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: oldPassword,
			},
		}, nil)
		if err != nil {
			err = api.ErrInvalidCredentials
			return
		}
	}

	changed, newInfo, err := ctx.Authenticators.UpdatePassword(oldInfo, &service.UpdatePasswordOptions{
		SetPassword:    true,
		PlainPassword:  newPassword,
		SetExpireAfter: true,
	})
	if err != nil {
		return
	}

	if !changed && e.Reason != nil && *e.Reason == interaction.AuthenticatorUpdateReasonExpiry {
		// Password is expired, but the user did not change the password.
		err = api.ErrPasswordReused
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
	Stage            authn.AuthenticationStage `json:"stage"`
	OldAuthenticator *authenticator.Info       `json:"old_authenticator"`
	NewAuthenticator *authenticator.Info       `json:"new_authenticator,omitempty"`
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
