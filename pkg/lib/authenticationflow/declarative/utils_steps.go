package declarative

import (
	"slices"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var authenticationStepTypes = []config.AuthenticationFlowStepType{
	config.AuthenticationFlowStepTypeAuthenticate,
	config.AuthenticationFlowStepTypeCreateAuthenticator,
	config.AuthenticationFlowStepTypeVerify,
}

// hasAuthenticationStep recursively checks if any of the steps in the obj are one of the specified authentication types.
func hasAuthenticationStep(obj config.AuthenticationFlowStepsObject, skip int) bool {
	steps := obj.GetSteps()
	if skip >= len(steps) {
		return false
	}

	for _, step := range steps[skip:] {
		flowStep, ok := step.(config.AuthenticationFlowObjectFlowStep)
		if !ok {
			// If it's not a flow step, it cannot be one of the types we are looking for.
			continue
		}

		stepType := flowStep.GetType()
		if slices.Contains(authenticationStepTypes, stepType) {
			return true
		}

		// Recursively check nested steps if available
		oneOfSteps := flowStep.GetOneOf()
		for _, oneOf := range oneOfSteps {
			if oneOfStepsObj, ok := oneOf.(config.AuthenticationFlowStepsObject); ok {
				// For nested calls, skip should be 0 as we are checking from the beginning of the nested steps.
				if hasAuthenticationStep(oneOfStepsObj, 0) {
					return true
				}
			}
		}
	}
	return false
}

func IsLastAuthentication(obj config.AuthenticationFlowStepsObject, skip int) bool {
	return !hasAuthenticationStep(obj, skip)
}
