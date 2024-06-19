package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type InlinePreviewAuthflowBranchViewModeler struct {
	AppConfig *config.AppConfig
}

func (m *InlinePreviewAuthflowBranchViewModeler) NewAuthflowBranchViewModelForInlinePreviewEnterPassword() AuthflowBranchViewModel {
	loginIDKeyType := model.LoginIDKeyTypeEmail
	if len(m.AppConfig.Identity.LoginID.Keys) > 0 {
		loginIDKeyType = m.AppConfig.Identity.LoginID.Keys[0].Type
	}
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
			OTPForm:          otp.FormCode,
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
