package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func authenticatorIsDefault(deps *authflow.Dependencies, userID string, authenticatorKind model.AuthenticatorKind) (isDefault bool, err error) {
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

func signupFlowCurrent(deps *authflow.Dependencies, id string, pointer jsonpointer.T) (config.AuthenticationFlowObject, error) {
	var root config.AuthenticationFlowObject

	if id == idGeneratedFlow {
		root = GenerateSignupFlowConfig(deps.Config)
	} else {
		for _, f := range deps.Config.AuthenticationFlow.SignupFlows {
			f := f
			if f.ID == id {
				root = f
				break
			}
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

func loginFlowCurrent(deps *authflow.Dependencies, id string, pointer jsonpointer.T) (config.AuthenticationFlowObject, error) {
	var root config.AuthenticationFlowObject

	if id == idGeneratedFlow {
		root = GenerateLoginFlowConfig(deps.Config)
	} else {
		for _, f := range deps.Config.AuthenticationFlow.LoginFlows {
			f := f
			if f.ID == id {
				root = f
				break
			}
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

func getAuthenticationCandidatesOfIdentity(deps *authflow.Dependencies, info *identity.Info, am config.AuthenticationFlowAuthentication) ([]UseAuthenticationCandidate, error) {
	as, err := deps.Authenticators.List(info.UserID, KeepAuthenticationMethod(am))
	if err != nil {
		return nil, err
	}

	return getAuthenticationCandidates(deps.Config.Authenticator.OOB, as, []config.AuthenticationFlowAuthentication{am})
}

func getAuthenticationCandidatesOfUser(deps *authflow.Dependencies, userID string, allAllowed []config.AuthenticationFlowAuthentication) ([]UseAuthenticationCandidate, error) {
	as, err := deps.Authenticators.List(userID, KeepAuthenticationMethod(allAllowed...))
	if err != nil {
		return nil, err
	}

	return getAuthenticationCandidates(deps.Config.Authenticator.OOB, as, allAllowed)
}

func getAuthenticationCandidatesForStep(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, userID string, step *config.AuthenticationFlowLoginFlowStep) ([]UseAuthenticationCandidate, error) {
	var candidates []UseAuthenticationCandidate

	for _, branch := range step.OneOf {
		switch branch.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			targetStepID := branch.TargetStep
			if targetStepID != "" {
				// Find the target step from the root.
				targetStepFlow, err := FindTargetStep(flows.Root, targetStepID)
				if err != nil {
					return nil, err
				}

				target, ok := targetStepFlow.Intent.(IntentLoginFlowStepAuthenticateTarget)
				if !ok {
					return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
						"target_step": targetStepID,
					})
				}

				identityInfo := target.GetIdentity(ctx, deps, flows.Replace(targetStepFlow))

				moreCandidates, err := getAuthenticationCandidatesOfIdentity(deps, identityInfo, branch.Authentication)
				if err != nil {
					return nil, err
				}

				candidates = append(candidates, moreCandidates...)
			} else {
				moreCandidates, err := getAuthenticationCandidatesOfUser(deps, userID, []config.AuthenticationFlowAuthentication{branch.Authentication})
				if err != nil {
					return nil, err
				}

				candidates = append(candidates, moreCandidates...)
			}
		case config.AuthenticationFlowAuthenticationDeviceToken:
			// Device token is handled transparently.
			break
		default:
			candidates = append(candidates, NewUseAuthenticationCandidateFromMethod(branch.Authentication))
		}
	}

	if len(candidates) == 0 {
		return nil, NoUsableAuthentication.New("no usable authentication method")
	}

	return candidates, nil
}

func getAuthenticationCandidates(oobConfig *config.AuthenticatorOOBConfig, as []*authenticator.Info, allAllowed []config.AuthenticationFlowAuthentication) (allUsable []UseAuthenticationCandidate, err error) {
	addOne := func() {
		added := false
		for _, a := range as {
			candidate := NewUseAuthenticationCandidateFromInfo(oobConfig, a)
			if !added {
				allUsable = append(allUsable, candidate)
				added = true
			}
		}
	}

	addAll := func() {
		for _, a := range as {
			candidate := NewUseAuthenticationCandidateFromInfo(oobConfig, a)
			allUsable = append(allUsable, candidate)
		}
	}

	for _, allowed := range allAllowed {
		switch allowed {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			addOne()
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			addAll()
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			addAll()
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			addOne()
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			addAll()
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			addAll()
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			addOne()
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			allUsable = append(allUsable, NewUseAuthenticationCandidateFromMethod(config.AuthenticationFlowAuthenticationRecoveryCode))
		case config.AuthenticationFlowAuthenticationDeviceToken:
			// Device token is handled transparently.
			break
		}
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

func getChannels(claimName model.ClaimName, oobConfig *config.AuthenticatorOOBConfig) []model.AuthenticatorOOBChannel {
	email := false
	sms := false
	whatsapp := false

	switch claimName {
	case model.ClaimEmail:
		email = true
	case model.ClaimPhoneNumber:
		switch oobConfig.SMS.PhoneOTPMode {
		case config.AuthenticatorPhoneOTPModeSMSOnly:
			sms = true
		case config.AuthenticatorPhoneOTPModeWhatsappOnly:
			whatsapp = true
		case config.AuthenticatorPhoneOTPModeWhatsappSMS:
			sms = true
			whatsapp = true
		}
	}

	channels := []model.AuthenticatorOOBChannel{}
	if email {
		channels = append(channels, model.AuthenticatorOOBChannelEmail)
	}
	if sms {
		channels = append(channels, model.AuthenticatorOOBChannelSMS)
	}
	if whatsapp {
		channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
	}

	return channels
}
