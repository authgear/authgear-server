package botprotection

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewCloudflareClient,
	NewRecaptchaV2Client,
)
