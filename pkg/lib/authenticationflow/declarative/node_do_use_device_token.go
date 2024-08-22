package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoUseDeviceToken{})
}

type NodeDoUseDeviceToken struct{}

var _ authflow.NodeSimple = &NodeDoUseDeviceToken{}
var _ authflow.Milestone = &NodeDoUseDeviceToken{}
var _ MilestoneDidSelectAuthenticationMethod = &NodeDoUseDeviceToken{}
var _ MilestoneDidAuthenticate = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) Kind() string {
	return "NodeDoUseDeviceToken"
}

func (*NodeDoUseDeviceToken) Milestone() {}
func (*NodeDoUseDeviceToken) MilestoneDidSelectAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return config.AuthenticationFlowAuthenticationDeviceToken
}
func (*NodeDoUseDeviceToken) MilestoneDidAuthenticate() (amr []string) { return }
