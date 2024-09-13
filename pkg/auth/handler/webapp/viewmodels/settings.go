package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SettingsViewModel struct {
	Authenticators           []*authenticator.Info
	NumberOfDeviceTokens     int
	HasDeviceTokens          bool
	ListRecoveryCodesAllowed bool
	ShowBiometric            bool

	HasSecondaryTOTP        bool
	HasSecondaryOOBOTPEmail bool
	HasSecondaryOOBOTPSMS   bool
	SecondaryPassword       *authenticator.Info
	HasMFA                  bool

	ShowSecondaryTOTP        bool
	ShowSecondaryOOBOTPEmail bool
	ShowSecondaryOOBOTPSMS   bool
	ShowSecondaryPassword    bool
	ShowMFA                  bool

	LatestPrimaryPasskey *authenticator.Info
	ShowPrimaryPasskey   bool
}

type SettingsIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type SettingsAuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsMFAService interface {
	CountDeviceTokens(userID string) (int, error)
}

type SettingsViewModeler struct {
	Authenticators SettingsAuthenticatorService
	MFA            SettingsMFAService
	Authentication *config.AuthenticationConfig
	Biometric      *config.BiometricConfig
}

// nolint: gocognit
func (m *SettingsViewModeler) ViewModel(userID string) (*SettingsViewModel, error) {
	authenticators, err := m.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}

	somePrimaryAuthenticatorCanHaveMFA := len(authenticator.ApplyFilters(
		authenticators,
		authenticator.KeepPrimaryAuthenticatorCanHaveMFA,
	)) > 0

	numberOfDeviceTokens, err := m.MFA.CountDeviceTokens(userID)
	if err != nil {
		return nil, err
	}
	hasDeviceTokens := numberOfDeviceTokens > 0

	hasSecondaryTOTP := false
	hasSecondaryOOBOTPEmail := false
	hasSecondaryOOBOTPSMS := false
	var secondaryPassword *authenticator.Info

	totpAllowed := false
	oobotpEmailAllowed := false
	oobotpSMSAllowed := false
	passwordAllowed := false
	passkeyAllowed := false

	var latestPrimaryPasskey *authenticator.Info

	if somePrimaryAuthenticatorCanHaveMFA {
		for _, typ := range *m.Authentication.SecondaryAuthenticators {
			switch typ {
			case model.AuthenticatorTypeTOTP:
				totpAllowed = true
			case model.AuthenticatorTypeOOBEmail:
				oobotpEmailAllowed = true
			case model.AuthenticatorTypeOOBSMS:
				oobotpSMSAllowed = true
			case model.AuthenticatorTypePassword:
				passwordAllowed = true
			}
		}
	}

	for _, typ := range *m.Authentication.PrimaryAuthenticators {
		switch typ {
		case model.AuthenticatorTypePasskey:
			passkeyAllowed = true
		}
	}

	for _, a := range authenticators {
		switch a.Kind {
		case authenticator.KindPrimary:
			switch a.Type {
			case model.AuthenticatorTypePasskey:
				aa := a
				latestPrimaryPasskey = aa
			}
		case authenticator.KindSecondary:
			switch a.Type {
			case model.AuthenticatorTypeTOTP:
				hasSecondaryTOTP = true
			case model.AuthenticatorTypeOOBEmail:
				hasSecondaryOOBOTPEmail = true
			case model.AuthenticatorTypeOOBSMS:
				hasSecondaryOOBOTPSMS = true
			case model.AuthenticatorTypePassword:
				aa := a
				secondaryPassword = aa
			}

		}
	}

	showBiometric := false
	for _, typ := range m.Authentication.Identities {
		if typ == model.IdentityTypeBiometric && *m.Biometric.ListEnabled {
			showBiometric = true
		}
	}

	hasMFA := (hasSecondaryTOTP ||
		hasSecondaryOOBOTPEmail ||
		hasSecondaryOOBOTPSMS ||
		secondaryPassword != nil)
	showSecondaryTOTP := hasSecondaryTOTP || totpAllowed
	showSecondaryOOBOTPEmail := hasSecondaryOOBOTPEmail || oobotpEmailAllowed
	showSecondaryOOBOTPSMS := hasSecondaryOOBOTPSMS || oobotpSMSAllowed
	showSecondaryPassword := secondaryPassword != nil || passwordAllowed
	showPrimaryPasskey := latestPrimaryPasskey != nil || passkeyAllowed
	showMFA := !m.Authentication.SecondaryAuthenticationMode.IsDisabled() &&
		(showSecondaryTOTP ||
			showSecondaryOOBOTPEmail ||
			showSecondaryOOBOTPSMS ||
			showSecondaryPassword)

	viewModel := &SettingsViewModel{
		Authenticators:           authenticators,
		NumberOfDeviceTokens:     numberOfDeviceTokens,
		HasDeviceTokens:          hasDeviceTokens,
		ListRecoveryCodesAllowed: !*m.Authentication.RecoveryCode.Disabled && m.Authentication.RecoveryCode.ListEnabled,
		ShowBiometric:            showBiometric,

		HasSecondaryTOTP:        hasSecondaryTOTP,
		HasSecondaryOOBOTPEmail: hasSecondaryOOBOTPEmail,
		HasSecondaryOOBOTPSMS:   hasSecondaryOOBOTPSMS,
		SecondaryPassword:       secondaryPassword,
		HasMFA:                  hasMFA,

		ShowSecondaryTOTP:        showSecondaryTOTP,
		ShowSecondaryOOBOTPEmail: showSecondaryOOBOTPEmail,
		ShowSecondaryOOBOTPSMS:   showSecondaryOOBOTPSMS,
		ShowSecondaryPassword:    showSecondaryPassword,
		ShowMFA:                  showMFA,

		LatestPrimaryPasskey: latestPrimaryPasskey,
		ShowPrimaryPasskey:   showPrimaryPasskey,
	}

	return viewModel, nil
}
