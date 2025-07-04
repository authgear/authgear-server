package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
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
func (*NodeDoUseDeviceToken) MilestoneDidSelectAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return model.AuthenticationFlowAuthenticationDeviceToken
}
func (*NodeDoUseDeviceToken) MilestoneDidAuthenticate() (amr []string) {
	return model.AuthenticationFlowAuthenticationDeviceToken.AMR()
}
func (*NodeDoUseDeviceToken) MilestoneDidAuthenticateAuthenticator() (*authenticator.Info, bool) {
	return nil, false
}
func (*NodeDoUseDeviceToken) MilestoneDidAuthenticateAuthentication() (*model.Authentication, bool) {
	return &model.Authentication{
		Authentication: model.AuthenticationFlowAuthenticationDeviceToken,
		Authenticator:  nil,
	}, true
}
