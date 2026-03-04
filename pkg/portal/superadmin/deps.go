package superadmin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/graphql"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	clock.DependencySet,
	globaldb.DependencySet,
	globalredis.DependencySet,
	graphql.DependencySet,
	transport.DependencySet,
)
