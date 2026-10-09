package config

import (
	"strconv"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type AuthenticationFlowForm struct {
	Name   string                         `json:"name,omitempty"`
	Fields []*AuthenticationFlowFormField `json:"fields,omitempty"`
}

// AuthenticationFlowFormField has exactly one non-nil member.
type AuthenticationFlowFormField struct {
	UserProfile *AuthenticationFlowFormFieldUserProfile `json:"user_profile,omitempty" nullable:"true"`
}

type AuthenticationFlowFormFieldUserProfile struct {
	Pointer  string `json:"pointer,omitempty"`
	Required bool   `json:"required,omitempty"`
}

// AuthenticationFlowObjectFormProvider is implemented by flow objects that can carry a form.
type AuthenticationFlowObjectFormProvider interface {
	AuthenticationFlowObject
	GetForm() *AuthenticationFlowForm
}

// validateAuthenticationFlowForms is called from AppConfig.Validate directly,
// not from validateAuthenticationFlow, which returns early when ui.authentication_flow is unset.
func validateAuthenticationFlowForms(ctx *validation.Context, c *AppConfig) {
	if c.AuthenticationFlow == nil {
		return
	}
	for i, f := range c.AuthenticationFlow.SignupFlows {
		validateFormsInSteps(ctx.Child("authentication_flow", "signup_flows", strconv.Itoa(i), "steps"), c.UserProfile, f.GetSteps(), false)
	}
	for i, f := range c.AuthenticationFlow.PromoteFlows {
		validateFormsInSteps(ctx.Child("authentication_flow", "promote_flows", strconv.Itoa(i), "steps"), c.UserProfile, f.GetSteps(), false)
	}
	for i, f := range c.AuthenticationFlow.LoginFlows {
		validateFormsInSteps(ctx.Child("authentication_flow", "login_flows", strconv.Itoa(i), "steps"), c.UserProfile, f.GetSteps(), false)
	}
}

// validateFormsInSteps tracks whether a user is identified before each step.
// Only an identify step changes that, so the result of a branch does not matter.
func validateFormsInSteps(ctx *validation.Context, up *UserProfileConfig, steps []AuthenticationFlowObject, identified bool) {
	for k, o := range steps {
		step, ok := o.(AuthenticationFlowObjectFlowStep)
		if !ok {
			continue
		}
		stepCtx := ctx.Child(strconv.Itoa(k))

		switch step.GetType() {
		case AuthenticationFlowStepTypeIdentify:
			validateFormsInBranches(stepCtx, up, step.GetOneOf(), true)
			identified = true
		case AuthenticationFlowStepTypeFillForm:
			if !identified {
				stepCtx.EmitErrorMessage("fill_form must come after an identify step")
			}
			if fp, ok := o.(AuthenticationFlowObjectFormProvider); ok && fp.GetForm() != nil {
				validateForm(stepCtx.Child("form"), up, fp.GetForm())
			}
		default:
			validateFormsInBranches(stepCtx, up, step.GetOneOf(), identified)
		}
	}
}

func validateFormsInBranches(ctx *validation.Context, up *UserProfileConfig, branches []AuthenticationFlowObject, identified bool) {
	for j, b := range branches {
		branchCtx := ctx.Child("one_of", strconv.Itoa(j))
		if fp, ok := b.(AuthenticationFlowObjectFormProvider); ok && fp.GetForm() != nil {
			validateForm(branchCtx.Child("form"), up, fp.GetForm())
		}
		if so, ok := b.(AuthenticationFlowStepsObject); ok {
			validateFormsInSteps(branchCtx.Child("steps"), up, so.GetSteps(), identified)
		}
	}
}

func validateForm(ctx *validation.Context, up *UserProfileConfig, form *AuthenticationFlowForm) {
	seen := map[string]struct{}{}
	for n, field := range form.Fields {
		if field.UserProfile == nil {
			continue
		}
		pointerCtx := ctx.Child("fields", strconv.Itoa(n), "user_profile", "pointer")
		pointer := field.UserProfile.Pointer
		if !isFillableUserProfilePointer(up, pointer) {
			pointerCtx.EmitErrorMessage("attribute is not fillable")
		}
		if _, ok := seen[pointer]; ok {
			pointerCtx.EmitErrorMessage("duplicated pointer")
		}
		seen[pointer] = struct{}{}
	}
}

// isFillableUserProfilePointer reports whether an end-user can fill the attribute in a form.
// It is evaluated against the current access control on every load and save.
func isFillableUserProfilePointer(up *UserProfileConfig, pointer string) bool {
	switch pointer {
	case "/email", "/phone_number", "/preferred_username":
		return false
	}
	if strings.HasPrefix(pointer, "/address/") {
		return false
	}

	if isStandardAttributePointer(pointer) {
		level := up.StandardAttributes.GetAccessControl().GetLevel(
			accesscontrol.Subject(pointer),
			RoleEndUser,
			AccessControlLevelHidden,
		)
		return level == AccessControlLevelReadwrite
	}

	for _, c := range up.CustomAttributes.Attributes {
		if c.Pointer == pointer {
			return c.AccessControl.EndUser == AccessControlLevelStringReadwrite
		}
	}
	return false
}

func isStandardAttributePointer(pointer string) bool {
	for _, p := range defaultReadwriteStandardAttributesPointers {
		if p == pointer {
			return true
		}
	}
	for _, p := range defaultHiddenStandardAttributesPointers {
		if p == pointer {
			return true
		}
	}
	return false
}
