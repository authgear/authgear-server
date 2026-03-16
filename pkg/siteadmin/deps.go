package siteadmin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/siteadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	clock.DependencySet,
	globaldb.DependencySet,
	globalredis.DependencySet,
	transport.DependencySet,
	wire.FieldsOf(new(*config.EnvironmentConfig), "CORSAllowedOrigins"),
	wire.Struct(new(CORSMatcher), "*"),
	wire.Bind(new(middleware.CORSOriginMatcher), new(*CORSMatcher)),
)
