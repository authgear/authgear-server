package dependency

import (
	"reflect"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type AuthDependency struct {
	dbProvider *db.DBProvider
}

// SetDBProvider set a DB provider implementation to dependency graph
func (d *AuthDependency) SetDBProvider(dbProvider db.DBProvider) {
	d.dbProvider = &dbProvider
}

func (d AuthDependency) Inject(h *handler.Handler, configuration config.TenantConfiguration) {
	t := reflect.TypeOf(h).Elem()
	v := reflect.ValueOf(h).Elem()

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		dependencyName := t.Field(i).Tag.Get("dependency")
		field := v.Field(i)
		field.Set(reflect.ValueOf(d.newGet(dependencyName, configuration)))
	}
}

func (d AuthDependency) newGet(dependencyName string, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "DB":
		return func() db.IDB {
			dbProvider := d.dbProvider
			if dbProvider == nil {
				return nil
			}

			return (*dbProvider).GetDB(tConfig)
		}
	default:
		return nil
	}
}
