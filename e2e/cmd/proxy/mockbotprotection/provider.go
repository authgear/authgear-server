package mockbotprotection

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Provider struct {
	// Authgear's type field for the provider
	Type config.BotProtectionProviderType
}

var ProviderCloudflare = Provider{
	Type: config.BotProtectionProviderTypeCloudflare,
}

var ProviderRecaptchaV2 = Provider{
	Type: config.BotProtectionProviderTypeRecaptchaV2,
}

var SupportedProviders = []Provider{
	ProviderCloudflare,
	ProviderRecaptchaV2,
}
