package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/phone"
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

func flowRootObject(deps *authflow.Dependencies, flowReference authflow.FlowReference) (config.AuthenticationFlowObject, error) {
	switch flowReference.Type {
	case authflow.FlowTypeSignup:
		return flowRootObjectForSignupFlow(deps, flowReference)
	case authflow.FlowTypeLogin:
		return flowRootObjectForLoginFlow(deps, flowReference)
	case authflow.FlowTypeSignupLogin:
		return flowRootObjectForSignupLoginFlow(deps, flowReference)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", flowReference.Type))
	}
}

func flowRootObjectForSignupFlow(deps *authflow.Dependencies, flowReference authflow.FlowReference) (config.AuthenticationFlowObject, error) {
	var root config.AuthenticationFlowObject

	if flowReference.Name == nameGeneratedFlow {
		root = GenerateSignupFlowConfig(deps.Config)
	} else {
		for _, f := range deps.Config.AuthenticationFlow.SignupFlows {
			f := f
			if f.Name == flowReference.Name {
				root = f
				break
			}
		}

	}

	if root == nil {
		return nil, ErrFlowNotFound
	}

	return root, nil
}

func flowRootObjectForLoginFlow(deps *authflow.Dependencies, flowReference authflow.FlowReference) (config.AuthenticationFlowObject, error) {
	var root config.AuthenticationFlowObject

	if flowReference.Name == nameGeneratedFlow {
		root = GenerateLoginFlowConfig(deps.Config)
	} else {
		for _, f := range deps.Config.AuthenticationFlow.LoginFlows {
			f := f
			if f.Name == flowReference.Name {
				root = f
				break
			}
		}
	}

	if root == nil {
		return nil, ErrFlowNotFound
	}

	return root, nil
}

func flowRootObjectForSignupLoginFlow(deps *authflow.Dependencies, flowReference authflow.FlowReference) (config.AuthenticationFlowObject, error) {
	var root config.AuthenticationFlowObject

	if flowReference.Name == nameGeneratedFlow {
		root = GenerateSignupLoginFlowConfig(deps.Config)
	} else {
		for _, f := range deps.Config.AuthenticationFlow.SignupLoginFlows {
			f := f
			if f.Name == flowReference.Name {
				root = f
				break
			}
		}
	}

	if root == nil {
		return nil, ErrFlowNotFound
	}

	return root, nil
}

func getAuthenticationOptionsForStep(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, userID string, step *config.AuthenticationFlowLoginFlowStep) ([]UseAuthenticationOption, error) {
	options := []UseAuthenticationOption{}

	infos, err := deps.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}

	recoveryCodes, err := deps.MFA.ListRecoveryCodes(userID)
	if err != nil {
		return nil, err
	}

	byTarget := func(am config.AuthenticationFlowAuthentication, targetStepName string) error {
		// Find the target step from the root.
		targetStepFlow, err := authflow.FindTargetStep(flows.Root, targetStepName)
		if err != nil {
			return err
		}

		target, ok := targetStepFlow.Intent.(IntentLoginFlowStepAuthenticateTarget)
		if !ok {
			return InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
				"target_step": targetStepName,
			})
		}

		identityInfo := target.GetIdentity(ctx, deps, flows.Replace(targetStepFlow))

		allAllowed := []config.AuthenticationFlowAuthentication{am}
		filteredInfos := authenticator.ApplyFilters(infos, KeepAuthenticationMethod(am), IsDependentOf(identityInfo))
		moreOptions, err := getAuthenticationOptions(deps, userID, filteredInfos, recoveryCodes, allAllowed)
		if err != nil {
			return err
		}

		options = append(options, moreOptions...)
		return nil
	}

	byUser := func(am config.AuthenticationFlowAuthentication) error {
		allAllowed := []config.AuthenticationFlowAuthentication{am}
		filteredInfos := authenticator.ApplyFilters(infos, KeepAuthenticationMethod(allAllowed...))
		moreOptions, err := getAuthenticationOptions(deps, userID, filteredInfos, recoveryCodes, allAllowed)
		if err != nil {
			return err
		}
		options = append(options, moreOptions...)
		return nil
	}

	for _, branch := range step.OneOf {
		switch branch.Authentication {
		case config.AuthenticationFlowAuthenticationDeviceToken:
			// Device token is handled transparently.
			break
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			err := byUser(branch.Authentication)
			if err != nil {
				return nil, err
			}
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			if targetStepName := branch.TargetStep; targetStepName != "" {
				err := byTarget(branch.Authentication, targetStepName)
				if err != nil {
					return nil, err
				}
			} else {
				err := byUser(branch.Authentication)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return options, nil
}

func getAuthenticationOptions(deps *authflow.Dependencies, userID string, as []*authenticator.Info, recoveryCodes []*mfa.RecoveryCode, allAllowed []config.AuthenticationFlowAuthentication) (allUsable []UseAuthenticationOption, err error) {
	addPrimaryPassword := func() {
		count := len(as)
		allUsable = append(allUsable, NewUseAuthenticationOptionPassword(
			config.AuthenticationFlowAuthenticationPrimaryPassword,
			count,
		))
	}

	addPasskeyIfPresent := func() error {
		if len(as) > 0 {
			requestOptions, err := deps.PasskeyRequestOptionsService.MakeModalRequestOptionsWithUser(userID)
			if err != nil {
				return err
			}

			allUsable = append(allUsable, NewUseAuthenticationOptionPasskey(requestOptions))
		}
		return nil
	}

	addSecondaryPasswordIfPresent := func() {
		count := len(as)
		if count > 0 {
			allUsable = append(allUsable, NewUseAuthenticationOptionPassword(
				config.AuthenticationFlowAuthenticationSecondaryPassword,
				count,
			))
		}
	}

	addTOTPIfPresent := func() {
		if len(as) > 0 {
			allUsable = append(allUsable, NewUseAuthenticationOptionTOTP())
		}
	}

	addAllOOBOTP := func() {
		for _, a := range as {
			option := NewUseAuthenticationOptionOOBOTP(deps.Config.Authenticator.OOB, a)
			allUsable = append(allUsable, option)
		}
	}

	addRecoveryCodeIfPresent := func() {
		if len(recoveryCodes) > 0 {
			allUsable = append(allUsable, NewUseAuthenticationOptionRecoveryCode())
		}
	}

	for _, allowed := range allAllowed {
		switch allowed {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			addPrimaryPassword()
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			err = addPasskeyIfPresent()
			if err != nil {
				return
			}
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			addSecondaryPasswordIfPresent()
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			addAllOOBOTP()
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			addTOTPIfPresent()
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			addRecoveryCodeIfPresent()
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
	channels := []model.AuthenticatorOOBChannel{}

	switch claimName {
	case model.ClaimEmail:
		channels = append(channels, model.AuthenticatorOOBChannelEmail)
	case model.ClaimPhoneNumber:
		switch oobConfig.SMS.PhoneOTPMode {
		case config.AuthenticatorPhoneOTPModeSMSOnly:
			channels = append(channels, model.AuthenticatorOOBChannelSMS)
		case config.AuthenticatorPhoneOTPModeWhatsappOnly:
			channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
		case config.AuthenticatorPhoneOTPModeWhatsappSMS:
			channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
			channels = append(channels, model.AuthenticatorOOBChannelSMS)
		}
	}

	return channels
}

func newIdentityInfo(deps *authflow.Dependencies, newUserID string, spec *identity.Spec) (*identity.Info, error) {
	// FIXME(authflow): allow bypassing email blocklist for Admin API.
	info, err := deps.Identities.New(newUserID, spec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	duplicate, err := deps.Identities.CheckDuplicated(info)
	if err != nil && !errors.Is(err, identity.ErrIdentityAlreadyExists) {
		return nil, err
	}

	if err != nil {
		spec := info.ToSpec()
		otherSpec := duplicate.ToSpec()
		return nil, identityFillDetails(api.ErrDuplicatedIdentity, &spec, &otherSpec)
	}

	return info, nil
}

func findExactOneIdentityInfo(deps *authflow.Dependencies, spec *identity.Spec) (*identity.Info, error) {
	bucketSpec := AccountEnumerationPerIPRateLimitBucketSpec(
		deps.Config.Authentication,
		string(deps.RemoteIP),
	)

	reservation := deps.RateLimiter.Reserve(bucketSpec)
	err := reservation.Error()
	if err != nil {
		return nil, err
	}
	defer deps.RateLimiter.Cancel(reservation)

	exactMatch, otherMatches, err := deps.Identities.SearchBySpec(spec)
	if err != nil {
		return nil, err
	}

	if exactMatch == nil {
		// Consume the reservation if exact match is not found.
		reservation.Consume()

		var otherSpec *identity.Spec
		if len(otherMatches) > 0 {
			s := otherMatches[0].ToSpec()
			otherSpec = &s
		}
		return nil, identityFillDetails(api.ErrUserNotFound, spec, otherSpec)
	}

	return exactMatch, nil
}

type HandleOAuthAuthorizationResponseOptions struct {
	Alias       string
	RedirectURI string
}

func handleOAuthAuthorizationResponse(deps *authflow.Dependencies, opts HandleOAuthAuthorizationResponseOptions, inputOAuth inputTakeOAuthAuthorizationResponse) (*identity.Spec, error) {
	if oauthError := inputOAuth.GetOAuthError(); oauthError != "" {
		errorDescription := inputOAuth.GetOAuthErrorDescription()
		errorURI := inputOAuth.GetOAuthErrorURI()

		return nil, sso.NewOAuthError(oauthError, errorDescription, errorURI)
	}

	oauthProvider := deps.OAuthProviderFactory.NewOAuthProvider(opts.Alias)
	if oauthProvider == nil {
		return nil, api.ErrOAuthProviderNotFound
	}

	code := inputOAuth.GetOAuthAuthorizationCode()

	// TODO(authflow): support nonce but do not save nonce in cookies.
	// Nonce in the current implementation is stored in cookies.
	// In the Authentication Flow API, cookies are not sent in Safari in third-party context.
	emptyNonce := ""
	authInfo, err := oauthProvider.GetAuthInfo(
		sso.OAuthAuthorizationResponse{
			Code: code,
		},
		sso.GetAuthInfoParam{
			RedirectURI: opts.RedirectURI,
			Nonce:       emptyNonce,
		},
	)
	if err != nil {
		return nil, err
	}

	providerConfig := oauthProvider.Config()
	providerID := providerConfig.ProviderID()
	identitySpec := &identity.Spec{
		Type: model.IdentityTypeOAuth,
		OAuth: &identity.OAuthSpec{
			ProviderID:     providerID,
			SubjectID:      authInfo.ProviderUserID,
			RawProfile:     authInfo.ProviderRawProfile,
			StandardClaims: authInfo.StandardAttributes.ToClaims(),
		},
	}

	return identitySpec, nil
}

type ConstructOAuthAuthorizationURLOptions struct {
	RedirectURI  string
	Alias        string
	ResponseMode sso.ResponseMode
}

func constructOAuthAuthorizationURL(ctx context.Context, deps *authflow.Dependencies, opts ConstructOAuthAuthorizationURLOptions) (authorizationURL string, err error) {
	oauthProvider := deps.OAuthProviderFactory.NewOAuthProvider(opts.Alias)
	if oauthProvider == nil {
		err = api.ErrOAuthProviderNotFound
		return
	}

	uiParam := uiparam.GetUIParam(ctx)

	param := sso.GetAuthURLParam{
		RedirectURI:  opts.RedirectURI,
		ResponseMode: opts.ResponseMode,
		Prompt:       uiParam.Prompt,
	}

	authorizationURL, err = oauthProvider.GetAuthURL(param)
	if err != nil {
		return
	}

	return
}

func getMaskedOTPTarget(claimName model.ClaimName, claimValue string) string {
	switch claimName {
	case model.ClaimEmail:
		return mail.MaskAddress(claimValue)
	case model.ClaimPhoneNumber:
		return phone.Mask(claimValue)
	default:
		panic(fmt.Errorf("unexpected claim name: %v", claimName))
	}
}
