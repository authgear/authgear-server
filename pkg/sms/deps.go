package sms

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewNexmoClient,
	NewTwilioClient,
	wire.Struct(new(ClientImpl), "*"),
)
