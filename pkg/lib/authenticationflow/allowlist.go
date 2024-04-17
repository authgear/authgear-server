package authenticationflow

import (
	"slices"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

// FlowAllowlist contains union of flow group and flow allowlist.
type FlowAllowlist []FlowAllowlistFlow

type FlowAllowlistFlow struct {
	Type FlowType `json:"type"`
	Name string   `json:"name"`
}

func NewFlowAllowlist(allowlist *config.AuthenticationFlowAllowlist, definedGroups []config.UIAuthenticationFlowGroup) FlowAllowlist {
	a := FlowAllowlist{}
	if allowlist == nil {
		return a
	}

	// Merge group allowlist
	if allowlist.Groups != nil {
		a = append(a, fromGroupAllowlist(*allowlist.Groups, definedGroups)...)
	}

	// Merge flow allowlist
	if allowlist.Flows != nil {
		a = append(a, fromFlowAllowlist(allowlist.Flows)...)
	}

	// Merge default group allowlist
	defaultGroupAllowlist := []string{"default"}
	a = append(a, fromGroupAllowlist(defaultGroupAllowlist, definedGroups)...)

	// Deduplicate
	a = slice.Deduplicate(a)

	return a
}

func (a FlowAllowlist) GetMostAppropriateFlowName(flowType FlowType) string {
	var flowName string
	for _, flow := range a {
		if flow.Type == flowType {
			flowName = flow.Name
			break
		}
	}

	if flowName == "" {
		flowName = "default"
	}

	return flowName
}

func (a FlowAllowlist) CanCreateFlow(flowReference FlowReference) bool {
	var flowNames []string
	for _, flow := range a {
		if flow.Type == flowReference.Type {
			flowNames = append(flowNames, flow.Name)
		}
	}

	// Allow all flows if the allowlist is not defined.
	if flowNames == nil {
		return true
	}

	return slices.Contains(flowNames, flowReference.Name)
}

func fromGroupAllowlist(groups []string, definedGroups []config.UIAuthenticationFlowGroup) FlowAllowlist {
	a := FlowAllowlist{}
	for _, groupName := range groups {
		for _, group := range definedGroups {
			if groupName == group.Name {
				a = append(a, fromGroupFlows(group.SignupFlows, FlowTypeSignup)...)
				a = append(a, fromGroupFlows(group.LoginFlows, FlowTypeLogin)...)
				a = append(a, fromGroupFlows(group.PromoteFlows, FlowTypePromote)...)
				a = append(a, fromGroupFlows(group.SignupLoginFlows, FlowTypeSignupLogin)...)
				a = append(a, fromGroupFlows(group.ReauthFlows, FlowTypeReauth)...)
				a = append(a, fromGroupFlows(group.AccountRecoveryFlows, FlowTypeAccountRecovery)...)
			}
		}
	}
	return a
}

func fromGroupFlows(flowNames []string, flowType FlowType) []FlowAllowlistFlow {
	allowlistFlows := []FlowAllowlistFlow{}
	for _, flowName := range flowNames {
		allowlistFlows = append(allowlistFlows, FlowAllowlistFlow{
			Type: flowType,
			Name: flowName,
		})
	}
	return allowlistFlows
}

func fromFlowAllowlist(flows *config.AuthenticationFlowAllowlistFlows) []FlowAllowlistFlow {
	allowlistFlows := []FlowAllowlistFlow{}
	if flows == nil {
		return allowlistFlows
	}

	for _, flow := range *flows {
		allowlistFlows = append(allowlistFlows, FlowAllowlistFlow{
			Type: FlowType(flow.Type),
			Name: flow.Name,
		})
	}
	return allowlistFlows
}
