package authflowv2

import (
	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
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

func validateBotProtectionInput(formData url.Values) error {
	return AuthflowBotProtectionSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(formData))
}

func insertBotProtection(formData url.Values, input map[string]interface{}) {
	bpType := formData.Get("x_bot_protection_provider_type")
	bpResp := formData.Get("x_bot_protection_provider_response")
	bot_protection := map[string]interface{}{
		"type":     bpType,
		"response": bpResp,
	}
	input["bot_protection"] = bot_protection
}

func HandleIdentificationBotProtection(identification config.AuthenticationFlowIdentification, flowResp *authflow.FlowResponse, formData url.Values, input map[string]interface{}) (err error) {
	bpRequired, err := webapp.IsIdentifyStepBotProtectionRequired(identification, flowResp)
	if err != nil {
		panic(err)
	}
	if bpRequired {
		err = validateBotProtectionInput(formData)
		if err != nil {
			return err
		}
		insertBotProtection(formData, input)
	}
	return
}

func HandleAccountRecoveryIdentificationBotProtection(identification config.AuthenticationFlowAccountRecoveryIdentification, flowResp *authflow.FlowResponse, formData url.Values, input map[string]interface{}) (err error) {
	bpRequired, err := webapp.IsAccountRecoveryIdentifyStepBotProtectionRequired(identification, flowResp)
	if err != nil {
		panic(err)
	}
	if bpRequired {
		err = validateBotProtectionInput(formData)
		if err != nil {
			return err
		}
		insertBotProtection(formData, input)
	}
	return
}
