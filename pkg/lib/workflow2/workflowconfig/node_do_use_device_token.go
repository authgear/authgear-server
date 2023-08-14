package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoUseDeviceToken{})
}

type NodeDoUseDeviceToken struct{}

var _ Milestone = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) Milestone() {}

var _ MilestoneAuthenticationMethod = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return config.WorkflowAuthenticationMethodDeviceToken
}

var _ MilestoneDidAuthenticate = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) MilestoneDidAuthenticate() (amr []string) { return }

var _ workflow.NodeSimple = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) Kind() string {
	return "workflowconfig.NodeDoUseDeviceToken"
}
