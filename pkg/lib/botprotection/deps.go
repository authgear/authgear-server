package botprotection

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Provider), "*"),
	NewCloudflareClient,
	NewRecaptchaV2Client,
)
