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
		return b.Authentication != model.AuthenticationFlowAuthenticationPrimaryPassword
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
		return b.Authentication != model.AuthenticationFlowAuthenticationPrimaryPasskey
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewEnterTOTP() AuthflowBranchViewModel {
	branches := m.generateAuthflowBranchesLoginIDAuthenticateSecondaries(m.getLoginIDKeyTypes())
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		return b.Authentication != model.AuthenticationFlowAuthenticationSecondaryTOTP
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
		return b.Authentication != model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
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
		return b.Authentication != model.AuthenticationFlowAuthenticationPrimaryPassword
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
	addedMap := make(map[model.AuthenticationFlowAuthentication]struct{})
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
	return reorderBranches(output)
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowBranchesIdentityLoginID(keyType model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch

	switch keyType {
	case model.LoginIDKeyTypeEmail:
		if branches, ok := m.generateSignupFlowBranchesAuthenticatePrimary(model.AuthenticationFlowIdentificationEmail); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypePhone:
		if branches, ok := m.generateSignupFlowBranchesAuthenticatePrimary(model.AuthenticationFlowIdentificationPhone); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypeUsername:
		if branches, ok := m.generateSignupFlowBranchesAuthenticatePrimary(model.AuthenticationFlowIdentificationUsername); ok {
			output = append(
				output,
				branches...,
			)
		}
	}

	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowBranchesAuthenticatePrimary(identification model.AuthenticationFlowIdentification) ([]AuthflowBranch, bool) {
	allowed := identification.PrimaryAuthentications()

	// This identification does not require primary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[model.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var output []AuthflowBranch

	for _, authenticatorType := range *m.AppConfig.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := model.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateSignupFlowStepAuthenticatePrimaryPassword()...)
			}
		case model.AuthenticatorTypeOOBEmail:
			am := model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateSignupFlowStepAuthenticatePrimaryOOBOTPEmail()...)
			}
		case model.AuthenticatorTypeOOBSMS:
			am := model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
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
			Authentication:        model.AuthenticationFlowAuthenticationPrimaryPassword,
			VerificationSkippable: true,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateSignupFlowStepAuthenticatePrimaryOOBOTPEmail() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:   model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
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
	channel := m.AppConfig.Authenticator.OOB.SMS.PhoneOTPMode.GetDefaultChannel()
	return []AuthflowBranch{
		{
			Authentication:        model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			Channel:               channel,
			MaskedClaimValue:      PreviewDummyPhoneNumberMasked,
			OTPForm:               otp.FormCode,
			VerificationSkippable: true,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesIdentityLoginIDs(keyTypes []model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch
	addedMap := make(map[model.AuthenticationFlowAuthentication]struct{})
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
	return reorderBranches(output)
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesIdentityLoginID(keyType model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch

	switch keyType {
	case model.LoginIDKeyTypeEmail:
		if branches, ok := m.generateAuthflowBranchesAuthenticatePrimary(model.AuthenticationFlowIdentificationEmail); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypePhone:
		if branches, ok := m.generateAuthflowBranchesAuthenticatePrimary(model.AuthenticationFlowIdentificationPhone); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypeUsername:
		if branches, ok := m.generateAuthflowBranchesAuthenticatePrimary(model.AuthenticationFlowIdentificationUsername); ok {
			output = append(
				output,
				branches...,
			)
		}
	}

	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesLoginIDAuthenticateSecondaries(keyTypes []model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch
	addedMap := make(map[model.AuthenticationFlowAuthentication]struct{})
	for _, typ := range keyTypes {
		branches := m.generateAuthflowBranchesLoginIDAuthenticateSecondary(typ)
		for _, branch := range branches {
			branch_ := branch
			if _, ok := addedMap[branch.Authentication]; !ok {
				addedMap[branch.Authentication] = struct{}{}
				output = append(output, branch_)
			}
		}
	}
	return reorderBranches(output)
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesLoginIDAuthenticateSecondary(keyType model.LoginIDKeyType) []AuthflowBranch {
	var output []AuthflowBranch

	switch keyType {
	case model.LoginIDKeyTypeEmail:
		if branches, ok := m.generateAuthflowBranchesAuthenticateSecondary(model.AuthenticationFlowIdentificationEmail); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypePhone:
		if branches, ok := m.generateAuthflowBranchesAuthenticateSecondary(model.AuthenticationFlowIdentificationPhone); ok {
			output = append(
				output,
				branches...,
			)
		}
	case model.LoginIDKeyTypeUsername:
		if branches, ok := m.generateAuthflowBranchesAuthenticateSecondary(model.AuthenticationFlowIdentificationUsername); ok {
			output = append(
				output,
				branches...,
			)
		}
	}

	return output
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesAuthenticatePrimary(identification model.AuthenticationFlowIdentification) ([]AuthflowBranch, bool) {
	allowed := identification.PrimaryAuthentications()

	// This identification does not require primary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[model.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var output []AuthflowBranch

	for _, authenticatorType := range *m.AppConfig.Authentication.PrimaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := model.AuthenticationFlowAuthenticationPrimaryPassword
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryPassword()...)
			}
		case model.AuthenticatorTypePasskey:
			am := model.AuthenticationFlowAuthenticationPrimaryPasskey
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryPasskey()...)
			}
		case model.AuthenticatorTypeOOBEmail:
			am := model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticatePrimaryOOBOTPEmail()...)
			}
		case model.AuthenticatorTypeOOBSMS:
			am := model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
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
			Authentication: model.AuthenticationFlowAuthenticationPrimaryPassword,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticatePrimaryPasskey() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: model.AuthenticationFlowAuthenticationPrimaryPasskey,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticatePrimaryOOBOTPEmail() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:   model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
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
	channel := m.AppConfig.Authenticator.OOB.SMS.PhoneOTPMode.GetDefaultChannel()
	return []AuthflowBranch{
		{
			Authentication:   model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			Channel:          channel,
			MaskedClaimValue: PreviewDummyPhoneNumberMasked,
			OTPForm:          otp.FormCode,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateAuthflowBranchesAuthenticateSecondary(identification model.AuthenticationFlowIdentification) ([]AuthflowBranch, bool) {
	if m.AppConfig.Authentication.SecondaryAuthenticationMode.IsDisabled() {
		return nil, false
	}
	allowed := identification.SecondaryAuthentications()

	// This identification does not require secondary authentication.
	if len(allowed) == 0 {
		return nil, false
	}

	allowedMap := make(map[model.AuthenticationFlowAuthentication]struct{})
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}

	var output []AuthflowBranch

	for _, authenticatorType := range *m.AppConfig.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			am := model.AuthenticationFlowAuthenticationSecondaryPassword
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryPassword()...)
			}
		case model.AuthenticatorTypeOOBEmail:
			am := model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryOOBOTPEmail()...)
			}
		case model.AuthenticatorTypeOOBSMS:
			am := model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS
			if _, ok := allowedMap[am]; ok {
				output = append(output, m.generateLoginFlowStepAuthenticateSecondaryOOBSMS()...)
			}
		case model.AuthenticatorTypeTOTP:
			am := model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS
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
			Authentication: model.AuthenticationFlowAuthenticationSecondaryPassword,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryOOBOTPEmail() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication:   model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
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
	channel := m.AppConfig.Authenticator.OOB.SMS.PhoneOTPMode.GetDefaultChannel()
	return []AuthflowBranch{
		{
			Authentication:   model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			Channel:          channel,
			MaskedClaimValue: PreviewDummyPhoneNumberMasked,
			OTPForm:          otp.FormCode,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryTOTP() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: model.AuthenticationFlowAuthenticationSecondaryTOTP,
		},
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) generateLoginFlowStepAuthenticateSecondaryRecoveryCode() []AuthflowBranch {
	return []AuthflowBranch{
		{
			Authentication: model.AuthenticationFlowAuthenticationRecoveryCode,
		},
	}
}
