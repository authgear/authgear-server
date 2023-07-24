//go:build wireinject
// +build wireinject

package service

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func newAuditSink(ctx context.Context, app *model.App, auditDatabase *auditdb.WriteHandle, loggerFactory *log.Factory) *audit.Sink {
	panic(wire.Build(AuthgearDependencySet))
}
