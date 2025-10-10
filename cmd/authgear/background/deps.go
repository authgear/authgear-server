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

	"github.com/authgear/authgear-server/pkg/lib/feature/accountanonymization"
	"github.com/authgear/authgear-server/pkg/lib/feature/accountdeletion"
	"github.com/authgear/authgear-server/pkg/lib/feature/accountstatus"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

// This dummy HTTP request is only used for get/set cookie
// which does not have any effect at all.
func NewDummyHTTPRequest() *http.Request {
	ctx := contextForDummyHTTPRequest
	r, _ := http.NewRequestWithContext(ctx, "", "", nil)
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

type AccountDeletionServiceFactory struct {
	BackgroundProvider *deps.BackgroundProvider
}

func (f *AccountDeletionServiceFactory) MakeUserService(appID string, appContext *config.AppContext) accountdeletion.UserService {
	return newUserService(f.BackgroundProvider, appID, appContext)
}

type AccountAnonymizationServiceFactory struct {
	BackgroundProvider *deps.BackgroundProvider
}

func (f *AccountAnonymizationServiceFactory) MakeUserService(appID string, appContext *config.AppContext) accountanonymization.UserService {
	return newUserService(f.BackgroundProvider, appID, appContext)
}

type AccountStatusServiceFactory struct {
	BackgroundProvider *deps.BackgroundProvider
}

func (f *AccountStatusServiceFactory) MakeUserService(appID string, appContext *config.AppContext) accountstatus.UserService {
	return newUserService(f.BackgroundProvider, appID, appContext)
}

type UserFacade interface {
	DeleteFromScheduledDeletion(ctx context.Context, userID string) error
	AnonymizeFromScheduledAnonymization(ctx context.Context, userID string) error
	RefreshAccountStatus(ctx context.Context, userID string) error
}

type UserService struct {
	AppDBHandle *appdb.Handle
	UserFacade  UserFacade
}

func (s *UserService) DeleteFromScheduledDeletion(ctx context.Context, userID string) (err error) {
	return s.AppDBHandle.WithTx(ctx, func(ctx context.Context) error {
		return s.UserFacade.DeleteFromScheduledDeletion(ctx, userID)
	})
}

func (s *UserService) AnonymizeFromScheduledAnonymization(ctx context.Context, userID string) (err error) {
	return s.AppDBHandle.WithTx(ctx, func(ctx context.Context) error {
		return s.UserFacade.AnonymizeFromScheduledAnonymization(ctx, userID)
	})
}

func (s *UserService) RefreshAccountStatus(ctx context.Context, userID string) (err error) {
	return s.AppDBHandle.WithTx(ctx, func(ctx context.Context) error {
		return s.UserFacade.RefreshAccountStatus(ctx, userID)
	})
}

var DependencySet = wire.NewSet(
	deps.BackgroundDependencySet,

	deps.CommonDependencySet,

	appdb.NewHandle,
	searchdb.NewHandle,
	appredis.NewHandle,
	globalredis.NewHandle,
	analyticredis.NewHandle,
	auditdb.NewReadHandle,
	auditdb.NewWriteHandle,
	NewDummyHTTPRequest,
	ProvideRemoteIP,
	ProvideUserAgentString,
	ProvideHTTPHost,
	ProvideHTTPProto,
	wire.Struct(new(AccountDeletionServiceFactory), "*"),
	wire.Struct(new(AccountAnonymizationServiceFactory), "*"),
	wire.Struct(new(AccountStatusServiceFactory), "*"),
	wire.Struct(new(UserService), "*"),
	wire.Bind(new(UserFacade), new(*facade.UserFacade)),
	wire.Bind(new(accountdeletion.UserServiceFactory), new(*AccountDeletionServiceFactory)),
	wire.Bind(new(accountanonymization.UserServiceFactory), new(*AccountAnonymizationServiceFactory)),
	wire.Bind(new(accountstatus.UserServiceFactory), new(*AccountStatusServiceFactory)),
	wire.Bind(new(event.Database), new(*appdb.Handle)),
	wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(loginid.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(hook.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(web.EmbeddedResourceManager), new(*web.GlobalEmbeddedResourceManager)),
)
