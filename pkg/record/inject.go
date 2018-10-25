package record

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
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

func openDB(tConfig config.TenantConfiguration) func() (*sqlx.DB, error) {
	return func() (*sqlx.DB, error) {
		return sqlx.Open("postgres", tConfig.DBConnectionStr)
	}
}

func (m DependencyMap) Provide(dependencyName string, r *http.Request) interface{} {
	switch dependencyName {
	case "AuthContextGetter":
		return coreAuth.NewContextGetterWithContext(r.Context())
	case "TxContext":
		tConfig := config.GetTenantConfig(r)
		return db.NewTxContextWithContext(r.Context(), openDB(tConfig))
	case "RecordStore":
		tConfig := config.GetTenantConfig(r)
		roleStore := auth.NewDefaultRoleStore(r.Context(), tConfig)
		return pq.NewRecordStore(
			roleStore,
			// TODO: get from tconfig
			true,
			db.NewSQLBuilder("record", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), openDB(tConfig))),
			logging.CreateLogger(r, "record", createLoggerMaskFormatter(r)),
		)
	default:
		return nil
	}
}

func createLoggerMaskFormatter(r *http.Request) logrus.Formatter {
	tConfig := config.GetTenantConfig(r)
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
