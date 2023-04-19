package portalapp

import (
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/web"
	portaldeps "github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	deps.CommonDependencySet,
	portaldeps.AppDependencySet,
	clock.DependencySet,
	globaldb.DependencySet,
	wire.Bind(new(EventService), new(*event.Service)),
	wire.Bind(new(identityloginid.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(hook.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.EmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),

	wire.Bind(new(event.Database), new(*appdb.Handle)),
	wire.NewSet(
		session.SessionUserIDGetterDependencySet,
		wire.Bind(new(event.SessionUserIDGetter), new(*session.SessionUserIDGetter)),
	),
)
