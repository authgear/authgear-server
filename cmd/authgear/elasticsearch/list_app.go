package elasticsearch

import (
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type AppLister struct {
	Handle *globaldb.Handle
	Store  *configsource.Store
}

func (l *AppLister) ListApps() (appIDs []string, err error) {
	err = l.Handle.ReadOnly(func() error {
		srcs, err := l.Store.ListAll()
		if err != nil {
			return err
		}
		for _, src := range srcs {
			appID := src.AppID
			appIDs = append(appIDs, appID)
		}
		return nil
	})
	if err != nil {
		return
	}
	return
}
