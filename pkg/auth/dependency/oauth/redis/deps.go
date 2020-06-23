package redis

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
)

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(GrantStore), "*"),
	wire.Bind(new(oauth.CodeGrantStore), new(*GrantStore)),
	wire.Bind(new(oauth.AccessGrantStore), new(*GrantStore)),
	wire.Bind(new(oauth.OfflineGrantStore), new(*GrantStore)),
)
