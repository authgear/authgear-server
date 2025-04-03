package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateAccountRecoveryFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowAccountRecoveryFlow {

	hasEmail := false
	hasPhone := false

	for _, key := range cfg.Identity.LoginID.Keys {
		switch key.Type {
		case model.LoginIDKeyTypeEmail:
			hasEmail = true

		case model.LoginIDKeyTypePhone:
			hasPhone = true
		}
	}

	oneOfs := []*config.AuthenticationFlowAccountRecoveryFlowOneOf{}
	if hasEmail {
		oneOfs = append(oneOfs, &config.AuthenticationFlowAccountRecoveryFlowOneOf{
			Identification:      config.AuthenticationFlowAccountRecoveryIdentificationEmail,
			OnFailure_WriteOnly: config.AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore,
			Steps: []*config.AuthenticationFlowAccountRecoveryFlowStep{
				{
					Type:            config.AuthenticationFlowAccountRecoveryFlowTypeSelectDestination,
					AllowedChannels: cfg.UI.ForgotPassword.Email,
				},
			},
		})
	}
	if hasPhone {
		oneOfs = append(oneOfs, &config.AuthenticationFlowAccountRecoveryFlowOneOf{
			Identification:      config.AuthenticationFlowAccountRecoveryIdentificationPhone,
			OnFailure_WriteOnly: config.AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore,
			Steps: []*config.AuthenticationFlowAccountRecoveryFlowStep{
				{
					Type:            config.AuthenticationFlowAccountRecoveryFlowTypeSelectDestination,
					AllowedChannels: cfg.UI.ForgotPassword.Phone,
				},
			},
		})
	}

	// Note we do not call getBotProtectionRequirementsOOBOTPEmail or getBotProtectionRequirementsOOBOTPSMS here.
	if bp, ok := getBotProtectionRequirementsAccountRecovery(cfg); ok {
		for _, oneOf := range oneOfs {
			oneOf.BotProtection = bp
		}
	}
	return &config.AuthenticationFlowAccountRecoveryFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowAccountRecoveryFlowStep{
			{
				Type:  config.AuthenticationFlowAccountRecoveryFlowTypeIdentify,
				OneOf: oneOfs,
			},
			{
				Type: config.AuthenticationFlowAccountRecoveryFlowTypeVerifyAccountRecoveryCode,
			},
			{
				Type: config.AuthenticationFlowAccountRecoveryFlowTypeResetPassword,
			},
		},
	}
}
