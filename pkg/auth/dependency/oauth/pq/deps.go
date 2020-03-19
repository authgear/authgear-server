package pq

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthorizationStore), "*"),
	wire.Bind(new(oauth.AuthorizationStore), new(*AuthorizationStore)),
)
