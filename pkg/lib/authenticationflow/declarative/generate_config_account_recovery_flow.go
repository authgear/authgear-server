package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateAccountRecoveryFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowAccountRecoveryFlow {
	return &config.AuthenticationFlowAccountRecoveryFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowAccountRecoveryFlowStep{
			{
				Type: config.AuthenticationFlowAccountRecoveryFlowTypeIdentify,
				OneOf: []*config.AuthenticationFlowAccountRecoveryFlowOneOf{
					{
						Identification: config.AuthenticationFlowAccountRecoveryIdentificationEmail,
						OnFailure:      config.AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore,
					},
					{
						Identification: config.AuthenticationFlowAccountRecoveryIdentificationPhone,
						OnFailure:      config.AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore,
					},
				},
			},
			{
				Type: config.AuthenticationFlowAccountRecoveryFlowTypeSelectDestination,
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
