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

func NewFlowAllowlist(allowlist *config.AuthenticationFlowAllowlist, definedGroups []*config.UIAuthenticationFlowGroup) FlowAllowlist {
	a := FlowAllowlist{}
	if allowlist == nil {
		return a
	}

	// Merge group allowlist
	a = append(a, fromGroupAllowlist(allowlist.Groups, definedGroups)...)

	// Merge flow allowlist
	a = append(a, fromFlowAllowlist(allowlist.Flows)...)

	// Merge default group allowlist
	defaultGroupAllowlist := []*config.AuthenticationFlowAllowlistGroup{{Name: "default"}}
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

func fromGroupAllowlist(groups []*config.AuthenticationFlowAllowlistGroup, definedGroups []*config.UIAuthenticationFlowGroup) FlowAllowlist {
	a := FlowAllowlist{}
	if groups == nil {
		return a
	}
	for _, group := range groups {
		for _, definedGroup := range definedGroups {
			if group.Name == definedGroup.Name {
				for _, flow := range definedGroup.Flows {
					a = append(a, FlowAllowlistFlow{
						Type: FlowType(flow.Type),
						Name: flow.Name,
					})
				}
				break
			}
		}
	}
	return a
}

func fromFlowAllowlist(flows []*config.AuthenticationFlowAllowlistFlow) []FlowAllowlistFlow {
	allowlistFlows := []FlowAllowlistFlow{}
	if flows == nil {
		return allowlistFlows
	}

	for _, flow := range flows {
		allowlistFlows = append(allowlistFlows, FlowAllowlistFlow{
			Type: FlowType(flow.Type),
			Name: flow.Name,
		})
	}
	return allowlistFlows
}
