package sms

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewNexmoClient,
	NewTwilioClient,
	NewCustomClient,
	NewLogger,
	NewHookHTTPClient,
	wire.Struct(new(Client), "*"),
	wire.Struct(new(SMSWebHook), "*"),
	wire.Struct(new(AntiSpamSMSBucketMaker), "*"),
)
