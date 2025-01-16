package sms

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
)

var DependencySet = wire.NewSet(
	NewLogger,
	custom.NewSMSHookTimeout,
	custom.NewHookHTTPClient,
	custom.NewHookDenoClient,
	wire.Struct(new(ClientResolver), "*"),
	wire.Struct(new(Sender), "*"),
	wire.Struct(new(custom.SMSWebHook), "*"),
	wire.Struct(new(custom.SMSDenoHook), "*"),
)
