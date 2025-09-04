package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoUpdatePreferredChannel{})
}

type NodeDoUpdatePreferredChannel struct {
	Info    *authenticator.Info           `json:"info,omitempty"`
	Channel model.AuthenticatorOOBChannel `json:"channel,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUpdatePreferredChannel{}
var _ authflow.EffectGetter = &NodeDoUpdatePreferredChannel{}
var _ authflow.Milestone = &NodeDoUpdatePreferredChannel{}
var _ MilestoneOOBOTPPreferredChannelUpdated = &NodeDoUpdatePreferredChannel{}

func (n *NodeDoUpdatePreferredChannel) Kind() string {
	return "NodeDoUpdatePreferredChannel"
}

func (n *NodeDoUpdatePreferredChannel) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			var channel string = string(n.Channel)
			n.Info.OOBOTP.PreferredChannel = &channel
			return deps.Authenticators.Update(ctx, n.Info)
		}),
	}, nil
}

func (*NodeDoUpdatePreferredChannel) Milestone()                              {}
func (*NodeDoUpdatePreferredChannel) MilestoneOOBOTPPreferredChannelUpdated() {}
