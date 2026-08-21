package configsource

import (
	"reflect"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// targetStepProvider is implemented by any flow object that can carry a target_step
// reference, at either the step level (verify, change_password) or the one_of branch
// level (create_authenticator, authenticate).
type targetStepProvider interface {
	GetTargetStepName() string
}

// validateAuthenticationFlowTargetSteps checks that every target_step refers to a step
// name that exists in the same flow.
//
// This runs on the write path only, so that saving a config with a dangling target_step
// is rejected while an app whose stored config already has one keeps loading and fails
// at runtime as before.
//
// Only flows that the incoming config actually changes are checked, so a dangling
// target_step that is already stored does not block edits to unrelated parts of the
// config.
//
// Only step existence is checked. Whether the target step is of a usable type
// (AuthenticationFlowInvalidTargetStep) and whether the flow is structurally coherent
// (AuthenticationFlowInvalidFlowConfig) are left to runtime, as both depend on intent
// types and runtime milestone state rather than on config structure.
//
// Only signup, promote and login flows are walked, because target_step exists nowhere
// else in the schema.
func (d AuthgearYAMLDescriptor) validateAuthenticationFlowTargetSteps(validationCtx *validation.Context, original *config.AppConfig, incoming *config.AppConfig) {
	if incoming.AuthenticationFlow == nil {
		return
	}

	var originalFlows config.AuthenticationFlowConfig
	if original != nil && original.AuthenticationFlow != nil {
		originalFlows = *original.AuthenticationFlow
	}

	// PromoteFlows is intentionally of type AuthenticationFlowSignupFlow.
	for i, flow := range incoming.AuthenticationFlow.SignupFlows {
		if isFlowUnchanged(originalFlows.SignupFlows, flow) {
			continue
		}
		validateTargetStepsInFlow(validationCtx, flow.GetSteps(), "signup_flows", i)
	}
	for i, flow := range incoming.AuthenticationFlow.PromoteFlows {
		if isFlowUnchanged(originalFlows.PromoteFlows, flow) {
			continue
		}
		validateTargetStepsInFlow(validationCtx, flow.GetSteps(), "promote_flows", i)
	}
	for i, flow := range incoming.AuthenticationFlow.LoginFlows {
		if isFlowUnchanged(originalFlows.LoginFlows, flow) {
			continue
		}
		validateTargetStepsInFlow(validationCtx, flow.GetSteps(), "login_flows", i)
	}
}

// isFlowUnchanged reports whether incoming is identical to the flow of the same name
// in originals. A flow with no counterpart in originals counts as changed, so a newly
// added flow is always checked.
//
// A target_step is resolved within a single flow, so an untouched flow cannot have
// acquired a dangling reference from an edit made elsewhere. Skipping it is what keeps
// an already stored dangling reference from blocking unrelated config updates.
func isFlowUnchanged[T interface{ GetName() string }](originals []T, incoming T) bool {
	for _, original := range originals {
		if original.GetName() == incoming.GetName() {
			return reflect.DeepEqual(original, incoming)
		}
	}
	return false
}

func validateTargetStepsInFlow(validationCtx *validation.Context, steps []config.AuthenticationFlowObject, flowsKey string, flowIndex int) {
	names := make(map[string]struct{})
	collectAuthenticationFlowStepNames(steps, names)
	basePath := []string{"authentication_flow", flowsKey, strconv.Itoa(flowIndex), "steps"}
	validateTargetStepRefs(validationCtx, names, steps, basePath)
}

// collectAuthenticationFlowStepNames collects every named step in the flow, including
// steps nested inside one_of branches. authenticationflow.FindTargetStep traverses the
// whole flow tree at runtime, so a target_step may name any step in the same flow.
func collectAuthenticationFlowStepNames(steps []config.AuthenticationFlowObject, out map[string]struct{}) {
	for _, stepObject := range steps {
		step, ok := stepObject.(config.AuthenticationFlowObjectFlowStep)
		if !ok {
			continue
		}
		if name := step.GetName(); name != "" {
			out[name] = struct{}{}
		}
		for _, branchObject := range step.GetOneOf() {
			if branch, ok := branchObject.(config.AuthenticationFlowStepsObject); ok {
				collectAuthenticationFlowStepNames(branch.GetSteps(), out)
			}
		}
	}
}

func validateTargetStepRefs(validationCtx *validation.Context, names map[string]struct{}, steps []config.AuthenticationFlowObject, path []string) {
	for i, stepObject := range steps {
		step, ok := stepObject.(config.AuthenticationFlowObjectFlowStep)
		if !ok {
			continue
		}
		stepPath := appendPath(path, strconv.Itoa(i))

		// Step-level target_step, used by verify and change_password.
		if provider, ok := stepObject.(targetStepProvider); ok {
			checkTargetStepRef(validationCtx, names, provider.GetTargetStepName(), appendPath(stepPath, "target_step"))
		}

		for j, branchObject := range step.GetOneOf() {
			branchPath := appendPath(stepPath, "one_of", strconv.Itoa(j))

			// Branch-level target_step, used by create_authenticator and authenticate.
			if provider, ok := branchObject.(targetStepProvider); ok {
				checkTargetStepRef(validationCtx, names, provider.GetTargetStepName(), appendPath(branchPath, "target_step"))
			}

			if branch, ok := branchObject.(config.AuthenticationFlowStepsObject); ok {
				validateTargetStepRefs(validationCtx, names, branch.GetSteps(), appendPath(branchPath, "steps"))
			}
		}
	}
}

func checkTargetStepRef(validationCtx *validation.Context, names map[string]struct{}, targetStep string, path []string) {
	if targetStep == "" {
		return
	}
	if _, ok := names[targetStep]; ok {
		return
	}
	validationCtx.Child(path...).EmitErrorMessage("target_step does not refer to any named step in the same flow")
}

// appendPath returns a newly allocated slice so that sibling recursive calls never
// share a backing array.
func appendPath(path []string, segments ...string) []string {
	out := make([]string, 0, len(path)+len(segments))
	out = append(out, path...)
	out = append(out, segments...)
	return out
}
