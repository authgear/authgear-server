package redis

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("oauth-grant-store")}
}

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(GrantStore), "*"),
	wire.Bind(new(oauth.CodeGrantStore), new(*GrantStore)),
	wire.Bind(new(oauth.AccessGrantStore), new(*GrantStore)),
	wire.Bind(new(oauth.OfflineGrantStore), new(*GrantStore)),
)
