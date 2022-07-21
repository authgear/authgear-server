package nodes

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
)

func init() {
	interaction.RegisterNode(&NodeForgotPasswordBegin{})
	interaction.RegisterNode(&NodeForgotPasswordEnd{})
}

type NodeForgotPasswordBegin struct {
	LoginIDKeys []config.LoginIDKeyConfig `json:"-"`
}

func (n *NodeForgotPasswordBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	n.LoginIDKeys = ctx.Config.Identity.LoginID.Keys
	return nil
}

func (n *NodeForgotPasswordBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeForgotPasswordBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{n.edge()}, nil
}

func (n *NodeForgotPasswordBegin) edge() *EdgeForgotPasswordSelectLoginID {
	return &EdgeForgotPasswordSelectLoginID{
		Configs: n.LoginIDKeys,
	}
}

func (n *NodeForgotPasswordBegin) GetIdentityCandidates() []identity.Candidate {
	return n.edge().GetIdentityCandidates()
}

type InputForgotPasswordSelectLoginID interface {
	GetLoginID() string
}

type EdgeForgotPasswordSelectLoginID struct {
	Configs     []config.LoginIDKeyConfig
	RedirectURI string
}

// GetIdentityCandidates implements IdentityCandidatesGetter.
func (e *EdgeForgotPasswordSelectLoginID) GetIdentityCandidates() []identity.Candidate {
	candidates := make([]identity.Candidate, len(e.Configs))
	for i, c := range e.Configs {
		conf := c
		candidates[i] = identity.NewLoginIDCandidate(&conf)
	}
	return candidates
}

func (e *EdgeForgotPasswordSelectLoginID) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputForgotPasswordSelectLoginID
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	loginID := input.GetLoginID()

	err := ctx.ForgotPassword.SendCode(loginID)
	if errors.Is(err, forgotpassword.ErrUserNotFound) {
		return nil, forgotpasswordFillDetails(interaction.ErrUserNotFound)
	}
	if err != nil {
		return nil, err
	}

	successPageCookie := ctx.CookieManager.ValueCookie(successpage.PathCookieDef, "/forgot_password/success")
	return &NodeForgotPasswordEnd{
		LoginID:           loginID,
		SuccessPageCookie: successPageCookie,
	}, nil
}

type NodeForgotPasswordEnd struct {
	LoginID           string       `json:"login_id"`
	SuccessPageCookie *http.Cookie `json:"success_page_cookie,omitempty"`
}

// GetCookies implements CookiesGetter
func (n *NodeForgotPasswordEnd) GetCookies() []*http.Cookie {
	if n.SuccessPageCookie == nil {
		return nil
	}
	return []*http.Cookie{n.SuccessPageCookie}
}

// GetLoginID implements ForgotPasswordSuccessNode.
func (n *NodeForgotPasswordEnd) GetLoginID() string {
	return n.LoginID
}

func (n *NodeForgotPasswordEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeForgotPasswordEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeForgotPasswordEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return nil, nil
}
