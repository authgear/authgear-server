package declarative

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestInputSchemaTakeIdentificationOptionIndex(t *testing.T) {
	Convey("InputSchemaTakeIdentificationOptionIndex", t, func() {
		test := func(s *InputSchemaTakeIdentificationOptionIndex, expected string) {
			b := s.SchemaBuilder()
			bytes, err := json.Marshal(b)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqualJSON, expected)
		}

		test(&InputSchemaTakeIdentificationOptionIndex{}, `
{
    "type": "object",
    "required": ["index"],
    "properties": {
        "index": { "type": "integer" }
    }
}
		`)

		dummyBotProtectionCfg := &config.BotProtectionConfig{
			Enabled: true,
			Provider: &config.BotProtectionProvider{
				Type: config.BotProtectionProviderTypeCloudflare,
			},
		}

		test(&InputSchemaTakeIdentificationOptionIndex{
			IsBotProtectionRequired: true,
			BotProtectionCfg:        dummyBotProtectionCfg,
		}, `
{
    "type": "object",
    "required": ["index", "bot_protection"],
    "properties": {
        "index": { "type": "integer" },
        "bot_protection": {
            "type": "object",
            "required": ["type", "response"],
            "properties": {
                "type": { "const": "cloudflare" },
                "response": { "type": "string" }
            }
        }
    }
}
		`)
	})
}
