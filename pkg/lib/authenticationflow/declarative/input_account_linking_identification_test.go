package declarative

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaAccountLinkingIdentificationCarriesBotProtection(t *testing.T) {
	Convey("InputSchemaAccountLinkingIdentification", t, func() {
		schema := &InputSchemaAccountLinkingIdentification{
			Options: []AccountLinkingIdentificationOptionInternal{
				{
					AccountLinkingIdentificationOption: AccountLinkingIdentificationOption{
						Identifcation: model.AuthenticationFlowIdentificationEmail,
					},
				},
			},
		}

		raw := json.RawMessage(`{
			"index": 0,
			"bot_protection": {
				"type": "cloudflare",
				"response": "token"
			}
		}`)

		input, err := schema.MakeInput(context.Background(), raw)
		So(err, ShouldBeNil)

		accountLinkingInput := input.(*InputAccountLinkingIdentification)
		So(accountLinkingInput.GetBotProtectionProviderType(), ShouldEqual, config.BotProtectionProviderTypeCloudflare)
		So(accountLinkingInput.GetBotProtectionProviderResponse(), ShouldEqual, "token")
	})
}

func TestSyntheticInputAccountLinkingIdentifyBotProtection(t *testing.T) {
	Convey("SyntheticInputAccountLinkingIdentify", t, func() {
		input := &SyntheticInputAccountLinkingIdentify{
			BotProtection: &InputTakeBotProtectionBody{
				Type:     config.BotProtectionProviderTypeCloudflare,
				Response: "token",
			},
		}

		So(input.GetBotProtectionProviderType(), ShouldEqual, config.BotProtectionProviderTypeCloudflare)
		So(input.GetBotProtectionProviderResponse(), ShouldEqual, "token")
	})
}
