//go:build wireinject
// +build wireinject

package service

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

func newAuditSink(
	app *model.App,
	pool *db.Pool,
	cfg *config.DatabaseEnvironmentConfig,
) *audit.Sink {
	panic(wire.Build(AuthgearDependencySet))
}

func newHookSink(app *model.App, denoEndpoint config.DenoEndpoint) *hook.Sink {
	panic(wire.Build(AuthgearDependencySet))
}
