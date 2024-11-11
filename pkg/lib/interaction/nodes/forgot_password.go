package nodes

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
)

func init() {
	interaction.RegisterNode(&NodeForgotPasswordBegin{})
	interaction.RegisterNode(&NodeForgotPasswordEnd{})
}

type EdgeForgotPasswordBegin struct {
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (e *EdgeForgotPasswordBegin) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeForgotPasswordBegin{
		IdentityInfo: e.IdentityInfo,
	}, nil
}

type NodeForgotPasswordBegin struct {
	LoginIDKeys  []config.LoginIDKeyConfig `json:"-"`
	IdentityInfo *identity.Info            `json:"identity_info"`
}

func (n *NodeForgotPasswordBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	n.LoginIDKeys = ctx.Config.Identity.LoginID.Keys
	return nil
}

func (n *NodeForgotPasswordBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeForgotPasswordBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{n.edge()}, nil
}

func (n *NodeForgotPasswordBegin) edge() *EdgeForgotPasswordSelectLoginID {
	return &EdgeForgotPasswordSelectLoginID{
		Configs:      n.LoginIDKeys,
		IdentityInfo: n.IdentityInfo,
	}
}

func (n *NodeForgotPasswordBegin) GetIdentityCandidates() []identity.Candidate {
	return n.edge().GetIdentityCandidates()
}

type EdgeForgotPasswordSelectLoginID struct {
	Configs      []config.LoginIDKeyConfig `json:"-"`
	RedirectURI  string                    `json:"-"`
	IdentityInfo *identity.Info            `json:"identity_info"`
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

func (e *EdgeForgotPasswordSelectLoginID) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	if e.IdentityInfo == nil {
		return nil, forgotpasswordFillDetails(api.ErrUserNotFound)
	}

	loginID := e.IdentityInfo.LoginID.LoginID

	err := ctx.ForgotPassword.SendCode(goCtx, loginID, nil)
	if errors.Is(err, forgotpassword.ErrUserNotFound) {
		return nil, forgotpasswordFillDetails(api.ErrUserNotFound)
	} else if apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore send code rate limits; show success to user anyways.
	} else if err != nil {
		return nil, err
	}

	successPageCookie := ctx.CookieManager.ValueCookie(successpage.PathCookieDef, "/flows/forgot_password/success")
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

func (n *NodeForgotPasswordEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeForgotPasswordEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeForgotPasswordEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return nil, nil
}
