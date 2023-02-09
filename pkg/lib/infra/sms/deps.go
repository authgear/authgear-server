package sms

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewNexmoClient,
	NewTwilioClient,
	NewCustomClient,
	NewLogger,
	NewSMSHookTimeout,
	NewHookHTTPClient,
	NewHookDenoClient,
	wire.Struct(new(Client), "*"),
	wire.Struct(new(SMSWebHook), "*"),
	wire.Struct(new(SMSDenoHook), "*"),
	wire.Struct(new(AntiSpamSMSBucketMaker), "*"),
)
