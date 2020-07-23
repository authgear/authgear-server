package nodes

import (
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeSelectIdentityAnonymous{})
}

type InputSelectIdentityAnonymous interface {
	GetAnonymousRequestToken() string
}

type EdgeSelectIdentityAnonymous struct {
}

func (e *EdgeSelectIdentityAnonymous) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityAnonymous)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	enabled := false
	for _, t := range ctx.Config.Authentication.Identities {
		if t == authn.IdentityTypeAnonymous {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, newinteraction.ConfigurationViolated.New("anonymous users are not allowed")
	}

	_, request, err := ctx.AnonymousIdentities.ParseRequest(input.GetAnonymousRequestToken())
	if err != nil {
		return nil, newinteraction.ErrInvalidCredentials
	}

	purpose, err := ctx.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeAnonymousRequest {
		return nil, newinteraction.ErrInvalidCredentials
	}

	panic("TODO(new_interaction): implements anonymous user signup/login")
}

type NodeSelectIdentityAnonymous struct {
	// FIXME: use key set instead of single key for anonymous identities
	Identity    *identity.Info          `json:"identity"`
	NewIdentity *identity.Info          `json:"new_identity"`
	KeySet      *jwk.Set                `json:"key_set"`
	Action      anonymous.RequestAction `json:"action"`
}

func (n *NodeSelectIdentityAnonymous) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	panic("implement me")
}

func (n *NodeSelectIdentityAnonymous) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	panic("implement me")
}
