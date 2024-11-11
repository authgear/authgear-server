package viewmodels

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type SettingsViewModel struct {
	Zoneinfo string

	Authenticators           []*authenticator.Info
	NumberOfDeviceTokens     int
	HasDeviceTokens          bool
	ListRecoveryCodesAllowed bool
	HasRecoveryCodes         bool
	ShowBiometric            bool

	HasSecondaryTOTP        bool
	HasSecondaryOOBOTPEmail bool
	HasSecondaryOOBOTPSMS   bool
	OOBOTPSMSDefaultChannel string
	SecondaryPassword       *authenticator.Info
	HasMFA                  bool
	PhoneOTPMode            string

	ShowSecondaryTOTP        bool
	ShowSecondaryOOBOTPEmail bool
	ShowSecondaryOOBOTPSMS   bool
	ShowSecondaryPassword    bool
	ShowMFA                  bool

	LatestPrimaryPasskey *authenticator.Info
	ShowPrimaryPasskey   bool
}

type SettingsUserService interface {
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type SettingsIdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}

type SettingsAuthenticatorService interface {
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsMFAService interface {
	CountDeviceTokens(ctx context.Context, userID string) (int, error)
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
}

type SettingsViewModeler struct {
	Clock               clock.Clock
	Users               SettingsUserService
	Authenticators      SettingsAuthenticatorService
	MFA                 SettingsMFAService
	AuthenticatorConfig *config.AuthenticatorConfig
	Authentication      *config.AuthenticationConfig
	Biometric           *config.BiometricConfig
}

// nolint: gocognit
func (m *SettingsViewModeler) ViewModel(ctx context.Context, userID string) (*SettingsViewModel, error) {
	user, err := m.Users.Get(ctx, userID, config.RoleEndUser)
	if err != nil {
		return nil, err
	}

	stdAttrs := user.StandardAttributes
	str := func(key string) string {
		value, _ := stdAttrs[key].(string)
		return value
	}

	zoneinfo := str(stdattrs.Zoneinfo)

	recoveryCodes, err := m.MFA.ListRecoveryCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	authenticators, err := m.Authenticators.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	somePrimaryAuthenticatorCanHaveMFA := len(authenticator.ApplyFilters(
		authenticators,
		authenticator.KeepPrimaryAuthenticatorCanHaveMFA,
	)) > 0

	numberOfDeviceTokens, err := m.MFA.CountDeviceTokens(ctx, userID)
	if err != nil {
		return nil, err
	}
	hasDeviceTokens := numberOfDeviceTokens > 0

	listRecoveryCodesAllowed := !*m.Authentication.RecoveryCode.Disabled && m.Authentication.RecoveryCode.ListEnabled
	hasRecoveryCodes := len(recoveryCodes) > 0

	hasSecondaryTOTP := false
	hasSecondaryOOBOTPEmail := false
	hasSecondaryOOBOTPSMS := false
	var secondaryPassword *authenticator.Info

	oobotpSMSDefaultChannel := m.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.GetDefaultChannel()
	phoneOTPMode := string(m.AuthenticatorConfig.OOB.SMS.PhoneOTPMode)

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
		Zoneinfo: zoneinfo,

		Authenticators:           authenticators,
		NumberOfDeviceTokens:     numberOfDeviceTokens,
		HasDeviceTokens:          hasDeviceTokens,
		ListRecoveryCodesAllowed: listRecoveryCodesAllowed,
		HasRecoveryCodes:         hasRecoveryCodes,
		ShowBiometric:            showBiometric,

		HasSecondaryTOTP:        hasSecondaryTOTP,
		HasSecondaryOOBOTPEmail: hasSecondaryOOBOTPEmail,
		HasSecondaryOOBOTPSMS:   hasSecondaryOOBOTPSMS,
		OOBOTPSMSDefaultChannel: string(oobotpSMSDefaultChannel),
		SecondaryPassword:       secondaryPassword,
		HasMFA:                  hasMFA,
		PhoneOTPMode:            phoneOTPMode,

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
