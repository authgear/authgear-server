package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityPasskey{})
}

type NodeDoUseIdentityPasskey struct {
	*NodeDoUseIdentity
	AssertionResponse []byte              `json:"assertion_response,omitempty"`
	Authenticator     *authenticator.Info `json:"authenticator,omitempty"`
	RequireUpdate     bool                `json:"require_update,omitempty"`
}

type NodeDoUseIdentityPasskeyOptions struct {
	Identity             *identity.Info
	IdentitySpec         *identity.Spec
	AssertionResponse    []byte
	Authenticator        *authenticator.Info
	RequireUpdate        bool
	RateLimitReservation *ratelimit.Reservation
}

func NewNodeDoUseIdentityPasskey(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts *NodeDoUseIdentityPasskeyOptions) (authenticationflow.ReactToResult, error) {
	nodeDoUseIden, err := NewNodeDoUseIdentity(ctx, deps, flows, &NodeDoUseIdentity{
		Identity:             opts.Identity,
		IdentitySpec:         opts.IdentitySpec,
		RateLimitReservation: opts.RateLimitReservation,
	})
	if err != nil {
		return nil, err
	}

	n := &NodeDoUseIdentityPasskey{
		NodeDoUseIdentity: nodeDoUseIden,
		AssertionResponse: opts.AssertionResponse,
		Authenticator:     opts.Authenticator,
		RequireUpdate:     opts.RequireUpdate,
	}

	n.NodeDoUseIdentity = nodeDoUseIden

	return authenticationflow.NewNodeSimple(n), nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentityPasskey{}
var _ authflow.EffectGetter = &NodeDoUseIdentityPasskey{}
var _ authflow.Milestone = &NodeDoUseIdentityPasskey{}
var _ authflow.InputReactor = &NodeDoUseIdentityPasskey{}
var _ MilestoneDoUseUser = &NodeDoUseIdentityPasskey{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentityPasskey{}
var _ MilestoneDidSelectAuthenticator = &NodeDoUseIdentityPasskey{}
var _ MilestoneDidAuthenticate = &NodeDoUseIdentityPasskey{}
var _ MilestoneGetIdentitySpecs = &NodeDoUseIdentityPasskey{}

func (*NodeDoUseIdentityPasskey) Kind() string {
	return "NodeDoUseIdentityPasskey"
}

func (n *NodeDoUseIdentityPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return n.NodeDoUseIdentity.CanReactTo(ctx, deps, flows)
}

func (n *NodeDoUseIdentityPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return n.NodeDoUseIdentity.ReactTo(ctx, deps, flows, input)
}

func (n *NodeDoUseIdentityPasskey) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.RequireUpdate {
				return deps.Authenticators.Update(ctx, n.Authenticator)
			}
			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.PasskeyService.ConsumeAssertionResponse(ctx, n.AssertionResponse)
		}),
	}, nil
}

func (*NodeDoUseIdentityPasskey) Milestone() {}
func (n *NodeDoUseIdentityPasskey) MilestoneDoUseUser() string {
	return n.Identity.UserID
}
func (n *NodeDoUseIdentityPasskey) MilestoneDoUseIdentity() *identity.Info {
	return n.NodeDoUseIdentity.MilestoneDoUseIdentity()
}
func (n *NodeDoUseIdentityPasskey) MilestoneDoUseIdentityIdentification() model.Identification {
	return n.NodeDoUseIdentity.MilestoneDoUseIdentityIdentification()
}
func (n *NodeDoUseIdentityPasskey) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoUseIdentityPasskey) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}

func (n *NodeDoUseIdentityPasskey) MilestoneGetIdentitySpecs() []*identity.Spec {
	return n.NodeDoUseIdentity.MilestoneGetIdentitySpecs()
}

func (n *NodeDoUseIdentityPasskey) MilestoneDidAuthenticateAuthenticator() (*authenticator.Info, bool) {
	return n.Authenticator, true
}
func (n *NodeDoUseIdentityPasskey) MilestoneDidAuthenticateAuthentication() (*model.Authentication, bool) {
	authn := n.Authenticator.ToAuthentication()
	authnModel := n.Authenticator.ToModel()
	return &model.Authentication{
		Authentication: authn,
		Authenticator:  &authnModel,
	}, true
}
