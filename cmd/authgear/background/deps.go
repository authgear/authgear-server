package background

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/accountdeletion"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type NoopTaskQueue struct{}

func (NoopTaskQueue) Enqueue(taskParam task.Param) {}

func NewNoopTaskQueue() NoopTaskQueue {
	return NoopTaskQueue{}
}

// This dummy HTTP request is only used for get/set cookie
// which does not have any effect at all.
func NewDummyHTTPRequest() *http.Request {
	r, _ := http.NewRequest("", "", nil)
	return r
}

func ProvideRemoteIP() httputil.RemoteIP {
	return "127.0.0.1"
}

func ProvideHTTPHost() httputil.HTTPHost {
	return ""
}

func ProvideHTTPProto() httputil.HTTPProto {
	return "http"
}

func ProvideUserAgentString() httputil.UserAgentString {
	return "authgear"
}

type UserServiceFactory struct {
	BackgroundProvider *deps.BackgroundProvider
}

func (f *UserServiceFactory) NewUserService(ctx context.Context, appID string, appContext *config.AppContext) accountdeletion.UserService {
	return newUserService(ctx, f.BackgroundProvider, appID, appContext)
}

type UserFacade interface {
	DeleteFromScheduledDeletion(userID string) error
}

type UserService struct {
	AppDBHandle *appdb.Handle
	UserFacade  UserFacade
}

func (s *UserService) DeleteFromScheduledDeletion(userID string) (err error) {
	return s.AppDBHandle.WithTx(func() error {
		return s.UserFacade.DeleteFromScheduledDeletion(userID)
	})
}

var DependencySet = wire.NewSet(
	deps.BackgroundDependencySet,

	deps.CommonDependencySet,

	appdb.NewHandle,
	appredis.NewHandle,
	auditdb.NewReadHandle,
	auditdb.NewWriteHandle,
	NewNoopTaskQueue,
	NewDummyHTTPRequest,
	ProvideRemoteIP,
	ProvideUserAgentString,
	ProvideHTTPHost,
	ProvideHTTPProto,
	wire.Struct(new(UserServiceFactory), "*"),
	wire.Struct(new(UserService), "*"),
	wire.Bind(new(UserFacade), new(*facade.UserFacade)),
	wire.Bind(new(accountdeletion.UserServiceFactory), new(*UserServiceFactory)),
	wire.Bind(new(task.Queue), new(NoopTaskQueue)),
	wire.Bind(new(event.Database), new(*appdb.Handle)),
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(loginid.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),
)
