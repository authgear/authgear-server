//go:build wireinject
// +build wireinject

package importer

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func NewImporter(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	appID config.AppID,
	loginIDEmailConfig *config.LoginIDEmailConfig,
) *Importer {
	panic(wire.Build(DependencySet))
}
