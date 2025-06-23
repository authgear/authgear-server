package webapp

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var AuthflowBotProtectionSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_bot_protection_provider_type": { "type": "string" },
			"x_bot_protection_provider_response": { "type": "string" }
		},
		"required": ["x_bot_protection_provider_type", "x_bot_protection_provider_response"]
	}
`)

func ValidateBotProtectionInput(ctx context.Context, formData url.Values) error {
	return AuthflowBotProtectionSchema.Validator().ValidateValue(ctx, FormToJSON(formData))
}

func IsBotProtectionInputValid(ctx context.Context, formData url.Values) bool {
	err := ValidateBotProtectionInput(ctx, formData)
	return err == nil
}

func InsertBotProtection(formData url.Values, input map[string]interface{}) {
	bpType := formData.Get("x_bot_protection_provider_type")
	bpResp := formData.Get("x_bot_protection_provider_response")
	bot_protection := map[string]interface{}{
		"type":     bpType,
		"response": bpResp,
	}
	input["bot_protection"] = bot_protection
}

func HandleIdentificationBotProtection(ctx context.Context, identification model.AuthenticationFlowIdentification, flowResp *authflow.FlowResponse, formData url.Values, input map[string]interface{}) (err error) {
	bpRequired, err := webapp.IsIdentifyStepBotProtectionRequired(identification, flowResp)
	if err != nil {
		panic(err)
	}
	if bpRequired {
		err = ValidateBotProtectionInput(ctx, formData)
		if err != nil {
			return err
		}
		InsertBotProtection(formData, input)
	}
	return
}

// As IntentAccountRecoveryFlowStepIdentify has it's own IdentificationData type to narrow down Identification as {"email", "phone"},
// we imitate the same logic in HandleIdentificationBotProtection here for account recovery
func HandleAccountRecoveryIdentificationBotProtection(ctx context.Context, identification config.AuthenticationFlowAccountRecoveryIdentification, flowResp *authflow.FlowResponse, formData url.Values, input map[string]interface{}) (err error) {
	bpRequired, err := webapp.IsAccountRecoveryIdentifyStepBotProtectionRequired(identification, flowResp)
	if err != nil {
		panic(err)
	}
	if bpRequired {
		err = ValidateBotProtectionInput(ctx, formData)
		if err != nil {
			return err
		}
		InsertBotProtection(formData, input)
	}
	return
}

func HandleAuthenticationBotProtection(ctx context.Context, authentication model.AuthenticationFlowAuthentication, flowResp *authflow.FlowResponse, formData url.Values, input map[string]interface{}) (err error) {
	bpRequired, err := webapp.IsAuthenticateStepBotProtectionRequired(authentication, flowResp)
	if err != nil {
		panic(err)
	}
	if bpRequired {
		err = ValidateBotProtectionInput(ctx, formData)
		if err != nil {
			return err
		}
		InsertBotProtection(formData, input)
	}
	return
}

func HandleCreateAuthenticatorBotProtection(ctx context.Context, authentication model.AuthenticationFlowAuthentication, flowResp *authflow.FlowResponse, formData url.Values, input map[string]interface{}) (err error) {
	bpRequired, err := webapp.IsCreateAuthenticatorStepBotProtectionRequired(authentication, flowResp)
	if err != nil {
		panic(err)
	}
	if bpRequired {
		err = ValidateBotProtectionInput(ctx, formData)
		if err != nil {
			return err
		}
		InsertBotProtection(formData, input)
	}
	return
}
