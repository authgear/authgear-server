package sms

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewNexmoClient,
	NewTwilioClient,
	NewLogger,
	wire.Struct(new(Client), "*"),
)
