package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoCreatePasskey{})
}

type NodeDoCreatePasskeyOptions struct {
	SkipCreate          bool
	Identity            *identity.Info
	Authenticator       *authenticator.Info
	AttestationResponse []byte
}

func NewNodeDoCreatePasskeyReactToResult(ctx context.Context, deps *authflow.Dependencies, opts NodeDoCreatePasskeyOptions) (authflow.ReactToResult, error) {
	nodeDoCreateIdentityOpts := NodeDoCreateIdentityOptions{
		SkipCreate: opts.SkipCreate,
		Identity:   opts.Identity,
	}
	doCreateIdenNode, delayedFunction, err := NewNodeDoCreateIdentity(ctx, deps, nodeDoCreateIdentityOpts)
	if err != nil {
		return nil, err
	}

	node := &NodeDoCreatePasskey{
		NodeDoCreateIdentity: doCreateIdenNode,
		SkipCreate:           opts.SkipCreate,
		Identity:             opts.Identity,
		Authenticator:        opts.Authenticator,
		AttestationResponse:  opts.AttestationResponse,
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node:                   authflow.NewNodeSimple(node),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

type NodeDoCreatePasskey struct {
	*NodeDoCreateIdentity
	SkipCreate          bool                `json:"skip_create,omitempty"`
	Identity            *identity.Info      `json:"identity,omitempty"`
	Authenticator       *authenticator.Info `json:"authenticator,omitempty"`
	AttestationResponse []byte              `json:"attestation_response,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreatePasskey{}
var _ authflow.EffectGetter = &NodeDoCreatePasskey{}
var _ authflow.Milestone = &NodeDoCreatePasskey{}
var _ authflow.InputReactor = &NodeDoCreatePasskey{}
var _ MilestoneDoCreateIdentity = &NodeDoCreatePasskey{}
var _ MilestoneDoCreateAuthenticator = &NodeDoCreatePasskey{}
var _ MilestoneDoCreatePasskey = &NodeDoCreatePasskey{}

func (n *NodeDoCreatePasskey) Kind() string {
	return "NodeDoCreatePasskey"
}

func (*NodeDoCreatePasskey) Milestone() {}
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentity() *identity.Info {
	return n.NodeDoCreateIdentity.MilestoneDoCreateIdentity()
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateAuthenticatorSkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info) {
	panic("NodeDoCreatePasskey does not support update authenticator")
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentitySkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreatePasskey) MilestoneDoCreateIdentityUpdate(newInfo *identity.Info) {
	panic("NodeDoCreatePasskey does not support update identity")
}
func (n *NodeDoCreatePasskey) MilestoneDoCreatePasskeyUpdateUserID(userID string) {
	n.NodeDoCreateIdentity.Identity = n.NodeDoCreateIdentity.Identity.UpdateUserID(userID)
	n.Authenticator = n.Authenticator.UpdateUserID(userID)
}

func (n *NodeDoCreatePasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return n.NodeDoCreateIdentity.CanReactTo(ctx, deps, flows)
}

func (n *NodeDoCreatePasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return n.NodeDoCreateIdentity.ReactTo(ctx, deps, flows, input)
}
func (n *NodeDoCreatePasskey) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	effects, err := n.NodeDoCreateIdentity.GetEffects(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	if n.SkipCreate {
		return effects, nil
	}

	newEffects := []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Create(ctx, n.Authenticator, false)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.PasskeyService.ConsumeAttestationResponse(ctx, n.AttestationResponse)
		}),
	}

	effects = append(effects, newEffects...)
	return effects, nil
}
