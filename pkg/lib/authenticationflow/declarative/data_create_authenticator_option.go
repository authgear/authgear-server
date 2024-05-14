package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type CreateAuthenticatorOption struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// PasswordPolicy is specific to primary_password and secondary_password.
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`

	// Target is specific to primary_oob_otp_email, primary_oob_otp_sms, secondary_oob_otp_email, secondary_oob_otp_sms.
	Target *CreateAuthenticatorTarget `json:"target,omitempty"`
}

type CreateAuthenticatorOptionInternal struct {
	CreateAuthenticatorOption
	UnmaskedTarget string
}

type CreateAuthenticatorTarget struct {
	MaskedDisplayName    string `json:"masked_display_name"`
	VerificationRequired bool   `json:"verification_required"`
}

func makeCreateAuthenticatorTarget(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	oneOf *config.AuthenticationFlowSignupFlowOneOf,
	userID string,
) (*CreateAuthenticatorTarget, string, error) {
	var target *CreateAuthenticatorTarget = nil
	var claimValue string
	var err error
	targetStep := oneOf.TargetStep
	if targetStep != "" {
		claimValue, err = getCreateAuthenticatorOOBOTPTargetFromTargetStep(ctx, deps, flows, targetStep)
		if err != nil {
			return nil, "", err
		}
		if claimValue == "" {
			return nil, "", nil
		}
		claimName := getOOBAuthenticatorType(oneOf.Authentication).ToClaimName()
		verified, err := getCreateAuthenticatorOOBOTPTargetVerified(deps, userID, claimName, claimValue)
		if err != nil {
			return nil, "", err
		}
		masked := ""
		switch claimName {
		case model.ClaimEmail:
			masked = mail.MaskAddress(claimValue)
		case model.ClaimPhoneNumber:
			masked = phone.Mask(claimValue)
		}
		target = &CreateAuthenticatorTarget{
			MaskedDisplayName:    masked,
			VerificationRequired: !verified && oneOf.IsVerificationRequired(),
		}
	}
	return target, claimValue, nil
}

func NewCreateAuthenticationOptions(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	step *config.AuthenticationFlowSignupFlowStep,
	userID string) ([]CreateAuthenticatorOptionInternal, error) {
	options := []CreateAuthenticatorOptionInternal{}
	passwordPolicy := NewPasswordPolicy(
		deps.FeatureConfig.Authenticator,
		deps.Config.Authenticator.Password.Policy,
	)
	for _, b := range step.OneOf {
		switch b.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			options = append(options, CreateAuthenticatorOptionInternal{
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.Authentication,
					PasswordPolicy: passwordPolicy,
				},
			})
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// Cannot create passkey in this step.
			break
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			target, unmaskedTarget, err := makeCreateAuthenticatorTarget(ctx, deps, flows, b, userID)
			if err != nil {
				return nil, err
			}
			if target == nil {
				// Skip this option, because the target step was skipped
				continue
			}
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimEmail, deps.Config.Authenticator.OOB)
			otpForm := getOTPForm(purpose, model.ClaimEmail, deps.Config.Authenticator.OOB.Email)
			options = append(options, CreateAuthenticatorOptionInternal{
				UnmaskedTarget: unmaskedTarget,
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.Authentication,
					OTPForm:        otpForm,
					Channels:       channels,
					Target:         target,
				},
			})
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			target, unmaskedTarget, err := makeCreateAuthenticatorTarget(ctx, deps, flows, b, userID)
			if err != nil {
				return nil, err
			}
			if target == nil {
				// Skip this option, because the target step was skipped
				continue
			}
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimPhoneNumber, deps.Config.Authenticator.OOB)
			otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, deps.Config.Authenticator.OOB.Email)
			options = append(options, CreateAuthenticatorOptionInternal{
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.Authentication,
					OTPForm:        otpForm,
					Channels:       channels,
					Target:         target,
				},
				UnmaskedTarget: unmaskedTarget,
			})
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			options = append(options, CreateAuthenticatorOptionInternal{
				CreateAuthenticatorOption: CreateAuthenticatorOption{
					Authentication: b.Authentication,
				},
			})
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			// Recovery code is not created in this step.
			break
		case config.AuthenticationFlowAuthenticationDeviceToken:
			// Device token is irrelevant in this step.
			break
		}
	}
	return options, nil
}
