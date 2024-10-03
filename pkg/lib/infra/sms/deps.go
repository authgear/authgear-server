package sms

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewLogger,
	NewSMSHookTimeout,
	NewHookHTTPClient,
	NewHookDenoClient,
	wire.Struct(new(ClientResolver), "*"),
	wire.Struct(new(Client), "*"),
	wire.Struct(new(SMSWebHook), "*"),
	wire.Struct(new(SMSDenoHook), "*"),
)
