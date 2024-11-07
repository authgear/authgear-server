package nodes

import (
	"context"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
)

func init() {
	interaction.RegisterNode(&NodeResetPasswordBegin{})
	interaction.RegisterNode(&NodeResetPasswordEnd{})
}

type NodeResetPasswordBegin struct{}

func (n *NodeResetPasswordBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeResetPasswordBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeResetPassword{}}, nil
}

type InputResetPassword interface {
	GetResetPasswordUserID() string
	GetNewPassword() string
	GeneratePassword() bool
	SendPassword() bool
	ChangeOnLogin() bool
}

type InputResetPasswordByCode interface {
	GetCode() string
	GetNewPassword() string
}

type EdgeResetPassword struct{}

func (e *EdgeResetPassword) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var resetInput InputResetPassword
	var codeInput InputResetPasswordByCode
	successPageCookie := ctx.CookieManager.ValueCookie(successpage.PathCookieDef, "/flows/reset_password/success")
	if interaction.Input(rawInput, &resetInput) {
		return &NodeResetPasswordEnd{
			InputResetPassword: resetInput,
			SuccessPageCookie:  successPageCookie,
		}, nil

	} else if interaction.Input(rawInput, &codeInput) {
		return &NodeResetPasswordEnd{
			InputResetPasswordByCode: codeInput,
			SuccessPageCookie:        successPageCookie,
		}, nil
	} else {
		return nil, interaction.ErrIncompatibleInput
	}
}

type NodeResetPasswordEnd struct {
	InputResetPassword       InputResetPassword       `json:"-"`
	InputResetPasswordByCode InputResetPasswordByCode `json:"-"`
	SuccessPageCookie        *http.Cookie             `json:"success_page_cookie,omitempty"`
}

// GetCookies implements CookiesGetter
func (n *NodeResetPasswordEnd) GetCookies() []*http.Cookie {
	if n.SuccessPageCookie == nil {
		return nil
	}
	return []*http.Cookie{n.SuccessPageCookie}
}

func (n *NodeResetPasswordEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.InputResetPassword != nil {
				resetInput := n.InputResetPassword

				userID := resetInput.GetResetPasswordUserID()
				newPassword := resetInput.GetNewPassword()
				generatePassword := resetInput.GeneratePassword()
				sendPassword := resetInput.SendPassword()

				var expireAfter *time.Time = nil
				now := ctx.Clock.NowUTC()
				if resetInput.ChangeOnLogin() {
					expireAfter = &now
				}

				var err error
				if generatePassword {
					newPassword, err = ctx.PasswordGenerator.Generate()
					if err != nil {
						return err
					}
				}

				err = ctx.ResetPassword.ChangePasswordByAdmin(goCtx, &forgotpassword.SetPasswordOptions{
					UserID:         userID,
					PlainPassword:  newPassword,
					SendPassword:   sendPassword,
					ExpireAfter:    expireAfter,
					SetExpireAfter: true,
				})
				if err != nil {
					return err
				}
			}

			if n.InputResetPasswordByCode != nil {
				codeInput := n.InputResetPasswordByCode
				code := codeInput.GetCode()
				newPassword := codeInput.GetNewPassword()

				err := ctx.ResetPassword.ResetPasswordByEndUser(goCtx, code, newPassword)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeResetPasswordEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
