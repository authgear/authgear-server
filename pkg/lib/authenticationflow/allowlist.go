package authenticationflow

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// FlowAllowlist contains union of flow group and flow allowlist.
type FlowAllowlist struct {
	DefinedGroups []*config.UIAuthenticationFlowGroup
	AllowedGroups []*config.AuthenticationFlowAllowlistGroup
	AllowedFlows  []*config.AuthenticationFlowAllowlistFlow
}

func NewFlowAllowlist(allowlist *config.AuthenticationFlowAllowlist, definedGroups []*config.UIAuthenticationFlowGroup) FlowAllowlist {
	a := FlowAllowlist{
		DefinedGroups: definedGroups,
	}

	if allowlist != nil {
		a.AllowedGroups = allowlist.Groups
	}

	if allowlist != nil {
		a.AllowedFlows = allowlist.Flows
	}

	return a
}

func (a FlowAllowlist) DeriveFlowNameForDefaultUI(flowType FlowType, flowGroup string) (string, error) {
	findFlowInDefinedGroups := func() (string, error) {
		for _, definedGroup := range a.DefinedGroups {
			if definedGroup.Name == flowGroup {
				for _, flowRef := range definedGroup.Flows {
					if FlowType(flowRef.Type) == flowType {
						return flowRef.Name, nil
					}
				}

				// If we reach here, the group is defined but it does not
				// contain the desired flow type.
				return "", ErrFlowNotAllowed
			}
		}

		// If we reach here, the group is undefined.
		// As a special case, if the undefined group is default, we return default.
		if flowGroup == "default" {
			return "default", nil
		}

		return "", ErrFlowNotAllowed
	}

	// The first step is to resolve flowGroup
	switch {
	case flowGroup == "" && len(a.AllowedGroups) == 0:
		flowGroup = "default"
		return findFlowInDefinedGroups()
	case flowGroup == "" && len(a.AllowedGroups) != 0:
		flowGroup = a.AllowedGroups[0].Name
		return findFlowInDefinedGroups()
	case flowGroup != "" && len(a.AllowedGroups) == 0:
		return findFlowInDefinedGroups()
	case flowGroup != "" && len(a.AllowedGroups) != 0:
		isAllowed := false
		for _, allowedGroup := range a.AllowedGroups {
			if allowedGroup.Name == flowGroup {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return "", ErrFlowNotAllowed
		}
		return findFlowInDefinedGroups()
	default:
		panic(fmt.Errorf("unreachable"))
	}
}

func (a FlowAllowlist) CanCreateFlow(flowReference FlowReference) bool {
	// If the allowlist is unspecified, then allow all flows.
	if len(a.AllowedGroups) == 0 && len(a.AllowedFlows) == 0 {
		return true
	}

	var effectiveAllowlist []*config.AuthenticationFlowAllowlistFlow

	effectiveAllowlist = append(effectiveAllowlist, a.AllowedFlows...)

	// For each allowed group,
	for _, allowedGroup := range a.AllowedGroups {
		isDefined := false
		for _, definedGroup := range a.DefinedGroups {
			// Look up the defined group.
			if allowedGroup.Name == definedGroup.Name {
				isDefined = true
				for _, flow := range definedGroup.Flows {
					effectiveAllowlist = append(effectiveAllowlist, &config.AuthenticationFlowAllowlistFlow{
						Type: flow.Type,
						Name: flow.Name,
					})
				}
			}
		}

		// As a special case, if the allowed group is default but it is not defined,
		// then the default group is assumed to be allow default flows.
		// So if the flow name is default, then it is allowed.
		if !isDefined && allowedGroup.Name == "default" {
			if flowReference.Name == "default" {
				return true
			}
		}
	}

	// Otherwise we consider the effective allowlist.
	for _, entry := range effectiveAllowlist {
		if FlowType(entry.Type) == flowReference.Type && entry.Name == flowReference.Name {
			return true
		}
	}

	return false
}
