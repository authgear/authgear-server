package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateIdentity{})
}

type NodeDoCreateIdentityOptions struct {
	SkipCreate   bool
	Identity     *identity.Info
	IdentitySpec *identity.Spec
}

func NewNodeDoCreateIdentity(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts NodeDoCreateIdentityOptions) (*NodeDoCreateIdentity, error) {
	n := &NodeDoCreateIdentity{
		SkipCreate:   opts.SkipCreate,
		Identity:     opts.Identity,
		IdentitySpec: opts.IdentitySpec,
	}

	return n, nil
}

func NewNodeDoCreateIdentityReactToResult(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts NodeDoCreateIdentityOptions) (authflow.ReactToResult, error) {
	node, err := NewNodeDoCreateIdentity(ctx, deps, flows, opts)
	if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(node), nil
}

type NodeDoCreateIdentity struct {
	SkipCreate   bool           `json:"skip_create,omitempty"`
	Identity     *identity.Info `json:"identity,omitempty"`
	IdentitySpec *identity.Spec `json:"identity_spec,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateIdentity{}
var _ authflow.Milestone = &NodeDoCreateIdentity{}
var _ MilestoneDoCreateIdentity = &NodeDoCreateIdentity{}
var _ MilestoneGetIdentitySpecs = &NodeDoCreateIdentity{}
var _ authflow.EffectGetter = &NodeDoCreateIdentity{}
var _ authflow.InputReactor = &NodeDoCreateIdentity{}

func (n *NodeDoCreateIdentity) Kind() string {
	return "NodeDoCreateIdentity"
}

func (*NodeDoCreateIdentity) Milestone() {}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentity() *identity.Info {
	return n.Identity
}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentityIdentification() model.Identification {
	return n.identification()
}
func (n *NodeDoCreateIdentity) MilestoneGetIdentitySpecs() []*identity.Spec {
	return []*identity.Spec{n.IdentitySpec}
}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentitySkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentityUpdate(newInfo *identity.Info) {
	n.Identity = newInfo
}

func (n *NodeDoCreateIdentity) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (n *NodeDoCreateIdentity) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return NewNodePostIdentified(ctx, deps, flows, &NodePostIdentifiedOptions{
		Identification: n.identification(),
	})
}

func (n *NodeDoCreateIdentity) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	if n.SkipCreate {
		return nil, nil
	}
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			err := deps.Identities.Create(ctx, n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateIdentity) identification() model.Identification {
	idmodel := n.Identity.ToModel()
	return model.Identification{
		Identification: n.Identity.ToIdentification(),
		Identity:       &idmodel,
		IDToken:        nil,
	}
}
