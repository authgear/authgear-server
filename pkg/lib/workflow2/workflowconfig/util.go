package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func authenticatorIsDefault(deps *workflow.Dependencies, userID string, authenticatorKind model.AuthenticatorKind) (isDefault bool, err error) {
	ais, err := deps.Authenticators.List(
		userID,
		authenticator.KeepKind(authenticatorKind),
		authenticator.KeepDefault,
	)
	if err != nil {
		return
	}

	isDefault = len(ais) == 0
	return
}

func signupFlowCurrent(deps *workflow.Dependencies, id string, pointer jsonpointer.T) (config.WorkflowObject, error) {
	var root config.WorkflowObject
	for _, f := range deps.Config.Workflow.SignupFlows {
		f := f
		if f.ID == id {
			root = f
			break
		}
	}
	if root == nil {
		return nil, ErrFlowNotFound
	}

	entries, err := Traverse(root, pointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func loginFlowCurrent(deps *workflow.Dependencies, id string, pointer jsonpointer.T) (config.WorkflowObject, error) {
	var root config.WorkflowObject
	for _, f := range deps.Config.Workflow.LoginFlows {
		f := f
		if f.ID == id {
			root = f
			break
		}
	}
	if root == nil {
		return nil, ErrFlowNotFound
	}

	entries, err := Traverse(root, pointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func getAuthenticationCandidatesOfIdentity(deps *workflow.Dependencies, info *identity.Info, am config.WorkflowAuthenticationMethod) ([]AuthenticationCandidate, error) {
	as, err := deps.Authenticators.List(info.UserID, KeepAuthenticationMethod(am))
	if err != nil {
		return nil, err
	}

	return getAuthenticationCandidates(as, []config.WorkflowAuthenticationMethod{am})
}

func getAuthenticationCandidatesOfUser(deps *workflow.Dependencies, userID string, allAllowed []config.WorkflowAuthenticationMethod) ([]AuthenticationCandidate, error) {
	as, err := deps.Authenticators.List(userID, KeepAuthenticationMethod(allAllowed...))
	if err != nil {
		return nil, err
	}

	return getAuthenticationCandidates(as, allAllowed)
}

func getAuthenticationCandidatesForStep(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, userID string, step *config.WorkflowLoginFlowStep) ([]AuthenticationCandidate, error) {
	var candidates []AuthenticationCandidate

	for _, branch := range step.OneOf {
		switch branch.Authentication {
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
			fallthrough
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
			targetStepID := branch.TargetStep
			if targetStepID != "" {
				// Find the target step from the root.
				targetStepWorkflow, err := FindTargetStep(workflows.Root, targetStepID)
				if err != nil {
					return nil, err
				}

				target, ok := targetStepWorkflow.Intent.(IntentLoginFlowStepAuthenticateTarget)
				if !ok {
					return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
						"target_step": targetStepID,
					})
				}

				identityInfo := target.GetIdentity(ctx, deps, workflows.Replace(targetStepWorkflow))

				moreCandidates, err := getAuthenticationCandidatesOfIdentity(deps, identityInfo, branch.Authentication)
				if err != nil {
					return nil, err
				}

				candidates = append(candidates, moreCandidates...)
			} else {
				moreCandidates, err := getAuthenticationCandidatesOfUser(deps, userID, []config.WorkflowAuthenticationMethod{branch.Authentication})
				if err != nil {
					return nil, err
				}

				candidates = append(candidates, moreCandidates...)
			}
		default:
			candidates = append(candidates, NewAuthenticationCandidateFromMethod(branch.Authentication))
		}
	}

	return candidates, nil
}

func getAuthenticationCandidates(as []*authenticator.Info, allAllowed []config.WorkflowAuthenticationMethod) (allUsable []AuthenticationCandidate, err error) {
	addOne := func() {
		added := false
		for _, a := range as {
			candidate := NewAuthenticationCandidateFromInfo(a)
			if !added {
				allUsable = append(allUsable, candidate)
				added = true
			}
		}
	}

	addAll := func() {
		for _, a := range as {
			candidate := NewAuthenticationCandidateFromInfo(a)
			allUsable = append(allUsable, candidate)
		}
	}

	for _, allowed := range allAllowed {
		switch allowed {
		case config.WorkflowAuthenticationMethodPrimaryPassword:
			addOne()
		case config.WorkflowAuthenticationMethodPrimaryPasskey:
			addOne()
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
			addAll()
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
			addAll()
		case config.WorkflowAuthenticationMethodSecondaryPassword:
			addOne()
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
			addAll()
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
			addAll()
		case config.WorkflowAuthenticationMethodSecondaryTOTP:
			addOne()
		case config.WorkflowAuthenticationMethodRecoveryCode:
			allUsable = append(allUsable, NewAuthenticationCandidateRecoveryCode())
		case config.WorkflowAuthenticationMethodDeviceToken:
			// Device token is handled transparently.
			break
		}
	}

	if len(allUsable) == 0 {
		err = NoUsableAuthenticationMethod.NewWithInfo("no usable authentication method", apierrors.Details{
			"allowed": allAllowed,
		})
		return
	}

	return
}

func identityFillDetails(err error, spec *identity.Spec, otherSpec *identity.Spec) error {
	details := errorutil.Details{}

	if spec != nil {
		details["IdentityTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.Type)
		switch spec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.OAuth.ProviderID.Type)
		}
	}

	if otherSpec != nil {
		details["IdentityTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.Type)
		switch otherSpec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.OAuth.ProviderID.Type)
		}
	}

	return errorutil.WithDetails(err, details)
}
