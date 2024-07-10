package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputTakeBotProtection(t *testing.T) {
	Convey("InputTakeBotProtectionBody", t, func() {
		dummyBotProtectionCfg := &config.BotProtectionConfig{
			Enabled: true,
			Provider: &config.BotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}
		test := func(expected string) {
			b := NewBotProtectionBodySchemaBuilder(dummyBotProtectionCfg)
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(`
{
    "properties": {
        "response": {
            "type": "string"
        },
        "type": {
            "const": "cloudflare"
        }
    },
    "required": [
        "type",
        "response"
    ],
    "type": "object"
}
`)
	})
	Convey("InputSchemaTakeBotProtection", t, func() {
		test := func(s *InputSchemaTakeBotProtection, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}
		var dummyBotProtectionCfg = &config.BotProtectionConfig{
			Enabled: true,
			Provider: &config.BotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}

		test(&InputSchemaTakeBotProtection{
			BotProtectionCfg: dummyBotProtectionCfg,
		}, `
{
    "properties": {
        "bot_protection": {
            "properties": {
                "response": {
                    "type": "string"
                },
                "type": {
                    "const": "cloudflare"
                }
            },
            "required": [
                "type",
                "response"
            ],
            "type": "object"
        }
    },
    "required": [
        "bot_protection"
    ],
    "type": "object"
}
        `)
	})
}
