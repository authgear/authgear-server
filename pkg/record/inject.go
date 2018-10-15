package record

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
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
	case "TxContext":
		tConfig := config.GetTenantConfig(r)
		return db.NewTxContextWithContext(r.Context(), openDB(tConfig))
	default:
		return nil
	}
}
