package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoUpdateLastUsedChannel{})
}

type NodeDoUpdateLastUsedChannel struct {
	Info    *authenticator.Info           `json:"info,omitempty"`
	Channel model.AuthenticatorOOBChannel `json:"channel,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUpdateLastUsedChannel{}
var _ authflow.EffectGetter = &NodeDoUpdateLastUsedChannel{}
var _ authflow.Milestone = &NodeDoUpdateLastUsedChannel{}
var _ MilestoneOOBOTPLastUsedChannelUpdated = &NodeDoUpdateLastUsedChannel{}

func (n *NodeDoUpdateLastUsedChannel) Kind() string {
	return "NodeDoUpdateLastUsedChannel"
}

func (n *NodeDoUpdateLastUsedChannel) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			n.Info.OOBOTP.SetLastUsedChannel(n.Channel)
			return deps.Authenticators.Update(ctx, n.Info)
		}),
	}, nil
}

func (*NodeDoUpdateLastUsedChannel) Milestone()                             {}
func (*NodeDoUpdateLastUsedChannel) MilestoneOOBOTPLastUsedChannelUpdated() {}
