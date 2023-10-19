package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GenerateRequestAccountRecoveryFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowRequestAccountRecoveryFlow {
	return &config.AuthenticationFlowRequestAccountRecoveryFlow{
		Name: nameGeneratedFlow,
		Steps: []*config.AuthenticationFlowRequestAccountRecoveryFlowStep{
			{
				Type: config.AuthenticationFlowAccountRecoveryFlowTypeIdentify,
				OneOf: []*config.AuthenticationFlowAccountRecoveryFlowOneOf{
					{
						Identification: config.AuthenticationFlowRequestAccountRecoveryIdentificationEmail,
						OnFailure:      config.AuthenticationFlowRequestAccountRecoveryIdentificationOnFailureIgnore,
					},
					{
						Identification: config.AuthenticationFlowRequestAccountRecoveryIdentificationPhone,
						OnFailure:      config.AuthenticationFlowRequestAccountRecoveryIdentificationOnFailureIgnore,
					},
				},
			},
			{
				Type: config.AuthenticationFlowAccountRecoveryFlowTypeSelectDestination,
			},
		},
	}
}
