package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const (
	defaultLoginIDKeyType = model.LoginIDKeyTypeEmail // We fake login id key type to be email if no login id is set
)

type InlinePreviewAuthflowBranchViewModeler struct {
	AppConfig *config.AppConfig
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewEnterPassword() AuthflowBranchViewModel {
	loginIDKeyType := m.getFirstLoginIDKeyType()
	branches := m.generateAuthflowBranchesIdentityLoginID(loginIDKeyType)
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
	loginIDKeyType := m.getFirstLoginIDKeyType()
	if loginIDKeyType == model.LoginIDKeyTypeUsername {
		loginIDKeyType = model.LoginIDKeyTypeEmail
	}
	branches := m.generateAuthflowBranchesIdentityLoginID(loginIDKeyType)
	branches = slice.Filter[AuthflowBranch](branches, func(b AuthflowBranch) bool {
		if loginIDKeyType == model.LoginIDKeyTypeEmail {
			return b.Authentication != config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
		}
		return b.Authentication != config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
	})
	return AuthflowBranchViewModel{
		FlowType:           authflow.FlowTypeLogin,
		ActionType:         authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate),
		DeviceTokenEnabled: false,
		Branches:           branches,
	}
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewUsePasskey() AuthflowBranchViewModel {
	loginIDKeyType := m.getFirstLoginIDKeyType()
	branches := m.generateAuthflowBranchesIdentityLoginID(loginIDKeyType)
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
	loginIDKeyType := m.getFirstLoginIDKeyType()
	branches := m.generateAuthflowBranchesLoginIDAuthenticateSecondary(loginIDKeyType)
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

func (m *InlinePreviewAuthflowBranchViewModeler) getFirstLoginIDKeyType() model.LoginIDKeyType {
	loginIDKeyType := defaultLoginIDKeyType
	if len(m.AppConfig.Identity.LoginID.Keys) > 0 {
		loginIDKeyType = m.AppConfig.Identity.LoginID.Keys[0].Type
	}
	return loginIDKeyType
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
			MaskedClaimValue: mail.MaskAddress(PreviewDummyEmail),
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
			MaskedClaimValue: phone.Mask(PreviewDummyPhoneNumber),
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
			MaskedClaimValue: mail.MaskAddress(PreviewDummyEmail),
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
			MaskedClaimValue: phone.Mask(PreviewDummyPhoneNumber),
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
