package record

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/asset/fs"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record/pq"
)

type DependencyMap struct{}

func NewDependencyMap() DependencyMap {
	return DependencyMap{}
}

// Provide provides dependency instance by name
// nolint: gocyclo, golint
func (m DependencyMap) Provide(
	dependencyName string,
	ctx context.Context,
	requestID string,
	tConfig config.TenantConfiguration,
) interface{} {
	switch dependencyName {
	case "AuthContextGetter":
		return coreAuth.NewContextGetterWithContext(ctx)
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	case "RecordStore":
		roleStore := auth.NewDefaultRoleStore(ctx, tConfig)
		return pq.NewSafeRecordStore(
			roleStore,
			// TODO: get from tconfig
			true,
			db.NewSQLBuilder("record", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "record", createLoggerMaskFormatter(tConfig)),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "HandlerLogger":
		return logging.CreateLoggerWithRequestID(requestID, "record", createLoggerMaskFormatter(tConfig))
	case "AssetStore":
		// TODO: get from tConfig
		return fs.NewAssetStore("", "", "", true, logging.CreateLoggerWithRequestID(requestID, "record", createLoggerMaskFormatter(tConfig)))
	default:
		return nil
	}
}

func createLoggerMaskFormatter(tConfig config.TenantConfiguration) logrus.Formatter {
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
