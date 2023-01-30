package sms

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewNexmoClient,
	NewTwilioClient,
	NewCustomClient,
	NewLogger,
	wire.Struct(new(Client), "*"),
	wire.Struct(new(AntiSpamSMSBucketMaker), "*"),
)
