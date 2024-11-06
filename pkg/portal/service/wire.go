//go:build wireinject
// +build wireinject

package service

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func newAuditSink(app *model.App, auditDatabase *auditdb.WriteHandle, loggerFactory *log.Factory) *audit.Sink {
	panic(wire.Build(AuthgearDependencySet))
}

func newHookSink(app *model.App, denoEndpoint config.DenoEndpoint, loggerFactory *log.Factory) *hook.Sink {
	panic(wire.Build(AuthgearDependencySet))
}
