package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/event"
)

func init() {
	newinteraction.RegisterNode(&NodeResetPasswordBegin{})
	newinteraction.RegisterNode(&NodeResetPasswordEnd{})
}

type NodeResetPasswordBegin struct{}

func (n *NodeResetPasswordBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeResetPasswordBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{&EdgeResetPassword{}}, nil
}

type InputResetPassword interface {
	GetCode() string
	GetNewPassword() string
}

type EdgeResetPassword struct{}

func (e *EdgeResetPassword) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputResetPassword)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	code := input.GetCode()
	newPassword := input.GetNewPassword()

	codeHash := ctx.ResetPassword.HashCode(code)
	userID, newInfo, updateInfo, err := ctx.ResetPassword.ResetPassword(code, newPassword)
	if err != nil {
		return nil, err
	}

	return &NodeResetPasswordEnd{
		UserID:              userID,
		NewAuthenticator:    newInfo,
		UpdateAuthenticator: updateInfo,
		CodeHash:            codeHash,
	}, nil
}

type NodeResetPasswordEnd struct {
	UserID              string              `json:"user_id"`
	NewAuthenticator    *authenticator.Info `json:"new_authenticator,omitempty"`
	UpdateAuthenticator *authenticator.Info `json:"update_authenticator,omitempty"`
	CodeHash            string              `json:"code_hash"`
}

func (n *NodeResetPasswordEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		if n.NewAuthenticator != nil {
			err := ctx.Authenticators.Create(n.NewAuthenticator)
			if err != nil {
				return err
			}
		}
		if n.UpdateAuthenticator != nil {
			err := ctx.Authenticators.Update(n.UpdateAuthenticator)
			if err != nil {
				return err
			}
		}

		user, err := ctx.Users.Get(n.UserID)
		if err != nil {
			return err
		}

		err = ctx.Hooks.DispatchEvent(
			event.PasswordUpdateEvent{
				Reason: event.PasswordUpdateReasonResetPassword,
				User:   *user,
			},
			user,
		)
		if err != nil {
			return err
		}

		err = ctx.ResetPassword.AfterResetPassword(n.CodeHash)
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeResetPasswordEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return nil, nil
}
