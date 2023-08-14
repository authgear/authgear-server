package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoUseDeviceToken{})
}

type NodeDoUseDeviceToken struct{}

var _ workflow.NodeSimple = &NodeDoUseDeviceToken{}
var _ workflow.Milestone = &NodeDoUseDeviceToken{}
var _ MilestoneAuthenticationMethod = &NodeDoUseDeviceToken{}
var _ MilestoneDidAuthenticate = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) Kind() string {
	return "workflowconfig.NodeDoUseDeviceToken"
}

func (*NodeDoUseDeviceToken) Milestone() {}
func (*NodeDoUseDeviceToken) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return config.WorkflowAuthenticationMethodDeviceToken
}
func (*NodeDoUseDeviceToken) MilestoneDidAuthenticate() (amr []string) { return }
