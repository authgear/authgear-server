package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const (
	defaultLoginIDKeyType = model.LoginIDKeyTypeEmail // We fake login id key type to be email if no login id is set
)

type InlinePreviewAuthflowBranchViewModeler struct {
	AppConfig *config.AppConfig
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewEnterPassword() AuthflowBranchViewModel {
	branches := m.generateAuthflowBranchesIdentityLoginIDs(m.getLoginIDKeyTypes())
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		return b.Authentication != config.AuthenticationFlowAuthenticationPrimaryPassword
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewEnterOOBOTP() AuthflowBranchViewModel {
	targetedLoginIDKeyTypes := m.getLoginIDKeyTypes()
	targetedLoginIDKeyTypes = slice.Filter(targetedLoginIDKeyTypes, func(t model.LoginIDKeyType) bool {
		return t != model.LoginIDKeyTypeUsername
	})
	branches := m.generateAuthflowBranchesIdentityLoginIDs(targetedLoginIDKeyTypes)
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewUsePasskey() AuthflowBranchViewModel {
	branches := m.generateAuthflowBranchesIdentityLoginIDs(m.getLoginIDKeyTypes())
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		return b.Authentication != config.AuthenticationFlowAuthenticationPrimaryPasskey
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewEnterTOTP() AuthflowBranchViewModel {
	branches := m.generateAuthflowBranchesIdentityLoginIDs(m.getLoginIDKeyTypes())
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		return b.Authentication != config.AuthenticationFlowAuthenticationSecondaryTOTP
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewOOBOTPLink() AuthflowBranchViewModel {
	branches := m.generateAuthflowBranchesIdentityLoginID(model.LoginIDKeyTypeEmail)
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		return b.Authentication != config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewCreatePassword() AuthflowBranchViewModel {
	branches := m.generateSignupFlowBranchesIdentityLoginIDs(m.getLoginIDKeyTypes())
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		return b.Authentication != config.AuthenticationFlowAuthenticationPrimaryPassword
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeSignup,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeCreateAuthenticator),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) getLoginIDKeyTypes() []model.LoginIDKeyType {
	if len(m.AppConfig.Identity.LoginID.Keys) == 0 {
		return []model.LoginIDKeyType{
			defaultLoginIDKeyType,
		}
	}
	return slice.Map(m.AppConfig.Identity.LoginID.Keys, func(key config.LoginIDKeyConfig) model.LoginIDKeyType {
		return key.Type
	})
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowBranchesIdentityLoginIDs(keyTypes []model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch
	addedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, typ := range keyTypes {
		branches := m.generateSignupFlowBranchesIdentityLoginID(typ)
		for _, branch := range branches {
			branch_ := branch
			if _, ok := addedMap[branch.Authentication]; !ok {
				addedMap[branch.Authentication] = struct{}{}
				output = append(output, branch_)
			}
		}
	}
	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowBranchesIdentityLoginID(keyType model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch

	switch keyType {
	case model.LoginIDKeyTypeEmail:
		if branches, ok := m.generateSignupFlowBranchesAuthenticatePrimary(config.AuthenticationFlowIdentificationEmail); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypePhone:
		if branches, ok := m.generateSignupFlowBranchesAuthenticatePrimary(config.AuthenticationFlowIdentificationPhone); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypeUsername:
		if branches, ok := m.generateSignupFlowBranchesAuthenticatePrimary(config.AuthenticationFlowIdentificationUsername); ok {
			output = append(
				output,
				branches...,
			)
		}
	}

	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowBranchesAuthenticatePrimary(identification config.AuthenticationFlowIdentification) ([]AuthflowBranch, bool) {
	allowed := identification.PrimaryAuthentications()

	// This identification does not require primary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var output []AuthflowBranch

	for _, authenticatorType := range *m.AppConfig.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := config.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateSignupFlowStepAuthenticatePrimaryPassword()...)
			}
		case model.AuthenticatorTypeOOBEmail:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateSignupFlowStepAuthenticatePrimaryOOBOTPEmail()...)
			}
		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateSignupFlowStepAuthenticatePrimaryOOBSMS()...)
			}
		}
	}

	return output, true
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowStepAuthenticatePrimaryPassword() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:        config.AuthenticationFlowAuthenticationPrimaryPassword,
			VerificationSkippable: true,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowStepAuthenticatePrimaryOOBOTPEmail() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:   config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
			Channel:          model.AuthenticatorOOBChannelEmail,
			MaskedClaimValue: PreviewDummyEmailMasked,
			OTPForm: func() otp.Form {
				if m.AppConfig.Authenticator.OOB.Email.EmailOTPMode.IsCodeEnabled() {
					return otp.FormCode
				}
				return otp.FormLink
			}(),
			VerificationSkippable: true,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowStepAuthenticatePrimaryOOBSMS() []AuthflowBranch {
	var channel model.AuthenticatorOOBChannel
	if m.AppConfig.Authenticator.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
		channel = model.AuthenticatorOOBChannelWhatsapp
	} else {
		channel = model.AuthenticatorOOBChannelSMS
	}
	return []AuthflowBranch{
		{
			Authentication:        config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			Channel:               channel,
			MaskedClaimValue:      PreviewDummyPhoneNumberMasked,
			OTPForm:               otp.FormCode,
			VerificationSkippable: true,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesIdentityLoginIDs(keyTypes []model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch
	addedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, typ := range keyTypes {
		branches := m.generateAuthflowBranchesIdentityLoginID(typ)
		for _, branch := range branches {
			branch_ := branch
			if _, ok := addedMap[branch.Authentication]; !ok {
				addedMap[branch.Authentication] = struct{}{}
				output = append(output, branch_)
			}
		}
	}
	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesIdentityLoginID(keyType model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch

	switch keyType {
	case model.LoginIDKeyTypeEmail:
		if branches, ok := m.generateAuthflowBranchesAuthenticatePrimary(config.AuthenticationFlowIdentificationEmail); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypePhone:
		if branches, ok := m.generateAuthflowBranchesAuthenticatePrimary(config.AuthenticationFlowIdentificationPhone); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypeUsername:
		if branches, ok := m.generateAuthflowBranchesAuthenticatePrimary(config.AuthenticationFlowIdentificationUsername); ok {
			output = append(
				output,
				branches...,
			)
		}
	}

	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesLoginIDAuthenticateSecondary(keyType model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch

	switch keyType {
	case model.LoginIDKeyTypeEmail:
		if branches, ok := m.generateAuthflowBranchesAuthenticateSecondary(config.AuthenticationFlowIdentificationEmail); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypePhone:
		if branches, ok := m.generateAuthflowBranchesAuthenticateSecondary(config.AuthenticationFlowIdentificationPhone); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypeUsername:
		if branches, ok := m.generateAuthflowBranchesAuthenticateSecondary(config.AuthenticationFlowIdentificationUsername); ok {
			output = append(
				output,
				branches...,
			)
		}
	}

	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesAuthenticatePrimary(identification config.AuthenticationFlowIdentification) ([]AuthflowBranch, bool) {
	allowed := identification.PrimaryAuthentications()

	// This identification does not require primary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var output []AuthflowBranch

	for _, authenticatorType := range *m.AppConfig.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := config.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryPassword()...)
			}
		case model.AuthenticatorTypePasskey:
			am := config.AuthenticationFlowAuthenticationPrimaryPasskey
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryPasskey()...)
			}
		case model.AuthenticatorTypeOOBEmail:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryOOBOTPEmail()...)
			}
		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryOOBSMS()...)
			}
		}
	}

	return output, true
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticatePrimaryPassword() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticatePrimaryPasskey() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticatePrimaryOOBOTPEmail() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:   config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
			Channel:          model.AuthenticatorOOBChannelEmail,
			MaskedClaimValue: PreviewDummyEmailMasked,
			OTPForm: func() otp.Form {
				if m.AppConfig.Authenticator.OOB.Email.EmailOTPMode.IsCodeEnabled() {
					return otp.FormCode
				}
				return otp.FormLink
			}(),
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticatePrimaryOOBSMS() []AuthflowBranch {
	var channel model.AuthenticatorOOBChannel
	if m.AppConfig.Authenticator.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
		channel = model.AuthenticatorOOBChannelWhatsapp
	} else {
		channel = model.AuthenticatorOOBChannelSMS
	}
	return []AuthflowBranch{
		{
			Authentication:   config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			Channel:          channel,
			MaskedClaimValue: PreviewDummyPhoneNumberMasked,
			OTPForm:          otp.FormCode,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesAuthenticateSecondary(identification config.AuthenticationFlowIdentification) ([]AuthflowBranch, bool) {
	if m.AppConfig.Authentication.SecondaryAuthenticationMode.IsDisabled() {
		return nil, false
	}
	allowed := identification.SecondaryAuthentications()

	// This identification does not require secondary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[config.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var output []AuthflowBranch

	for _, authenticatorType := range *m.AppConfig.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := config.AuthenticationFlowAuthenticationSecondaryPassword
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryPassword()...)
			}
		case model.AuthenticatorTypeOOBEmail:
			am := config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryOOBOTPEmail()...)
			}
		case model.AuthenticatorTypeOOBSMS:
			am := config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryOOBSMS()...)
			}
		case model.AuthenticatorTypeTOTP:
			am := config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryTOTP()...)
			}
		}
	}

	if !*m.AppConfig.Authentication.RecoveryCode.Disabled {
		output = append(output, m.generateLoginFlowStepAuthenticateSecondaryRecoveryCode()...)
	}

	return output, true
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryPassword() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryOOBOTPEmail() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:   config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
			Channel:          model.AuthenticatorOOBChannelEmail,
			MaskedClaimValue: PreviewDummyEmailMasked,
			OTPForm: func() otp.Form {
				if m.AppConfig.Authenticator.OOB.Email.EmailOTPMode.IsCodeEnabled() {
					return otp.FormCode
				}
				return otp.FormLink
			}(),
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryOOBSMS() []AuthflowBranch {
	var channel model.AuthenticatorOOBChannel
	if m.AppConfig.Authenticator.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
		channel = model.AuthenticatorOOBChannelWhatsapp
	} else {
		channel = model.AuthenticatorOOBChannelSMS
	}
	return []AuthflowBranch{
		{
			Authentication:   config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			Channel:          channel,
			MaskedClaimValue: PreviewDummyPhoneNumberMasked,
			OTPForm:          otp.FormCode,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryTOTP() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryRecoveryCode() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
		},
	}
}
