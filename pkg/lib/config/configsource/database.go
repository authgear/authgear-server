package configsource

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type DatabaseSource struct {
	ID        string
	AppID     string
	Data      map[string][]byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DatabaseLogger struct{ *log.Logger }

func NewDatabaseLogger(lf *log.Factory) DatabaseLogger {
	return DatabaseLogger{lf.New("configsource-database")}
}

type Database struct {
	Logger        DatabaseLogger
	BaseResources *resource.Manager
	TrustProxy    config.TrustProxy
	Config        *Config
	Clock         clock.Clock

	done     chan<- struct{} `wire:"-"`
	Store    *Store
	Database *globaldb.Handle
}

func (d *Database) Open() error {
	return nil
}

func (d *Database) Close() error {
	return nil
}

func (d *Database) ResolveAppID(r *http.Request) (string, error) {
	host := httputil.GetHost(r, bool(d.TrustProxy))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	var appID string
	err := d.Database.WithTx(func() error {
		aid, err := d.Store.GetAppIDByDomain(host)
		if err != nil {
			return err
		}
		appID = aid
		return nil
	})

	if err != nil {
		return "", err
	}
	return appID, nil
}

func (d *Database) ResolveContext(appID string) (*config.AppContext, error) {
	// fixme(1127) cache
	app := &dbApp{
		appID: appID,
	}

	var appCtx *config.AppContext
	err := d.Database.WithTx(func() error {
		a, err := app.Load(d)
		if err != nil {
			return err
		}
		appCtx = a
		return nil
	})
	return appCtx, err
}

func (d *Database) ReloadApp(appID string) {
}

func (d *Database) CreateDatabaseSource(appID string, resources map[string][]byte) error {
	return d.Database.WithTx(func() error {
		_, err := d.Store.GetDatabaseSourceByAppID(appID)
		if err != nil && !errors.Is(err, ErrAppNotFound) {
			return err
		} else if err == nil {
			return ErrDuplicatedAppID
		}

		dbData := make(map[string][]byte)
		for path, data := range resources {
			dbData[EscapePath(path)] = data
		}

		dbSource := &DatabaseSource{
			ID:        uuid.New(),
			AppID:     appID,
			Data:      dbData,
			CreatedAt: d.Clock.NowUTC(),
			UpdatedAt: d.Clock.NowUTC(),
		}
		return d.Store.CreateDatabaseSource(dbSource)
	})
}

func (d *Database) UpdateDatabaseSource(appID string, updates []*resource.ResourceFile) error {
	return d.Database.WithTx(func() error {
		dbs, err := d.Store.GetDatabaseSourceByAppID(appID)
		if err != nil {
			return err
		}

		updated := false
		for _, u := range updates {
			key := EscapePath(u.Location.Path)
			if u.Data == nil {
				if _, ok := dbs.Data[key]; ok {
					delete(dbs.Data, key)
					updated = true
				}
			} else {
				if !bytes.Equal(dbs.Data[key], u.Data) {
					dbs.Data[key] = u.Data
					updated = true
				}
			}
		}

		if updated {
			dbs.UpdatedAt = d.Clock.NowUTC()
			err = d.Store.UpdateDatabaseSource(dbs)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func MakeAppFSFromDatabaseSource(s *DatabaseSource) (resource.Fs, error) {
	// Construct a FS that treats `a` and `/a` the same.
	// The template is loaded by a file URI which is always an absoluted path.
	appFs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	create := func(name string, data []byte) {
		file, _ := appFs.Create(name)
		_, _ = file.Write(data)
	}

	for key, data := range s.Data {
		path, err := UnescapePath(key)
		if err != nil {
			return nil, err
		}
		create(path, data)
	}

	return &resource.LeveledAferoFs{
		Fs:      appFs,
		FsLevel: resource.FsLevelApp,
	}, nil
}

type dbApp struct {
	appID      string
	appCtx     *config.AppContext
	err        error
	lastUsedAt int64
}

func (a *dbApp) Load(d *Database) (*config.AppContext, error) {
	a.appCtx, a.err = a.doLoad(d)
	atomic.StoreInt64(&a.lastUsedAt, d.Clock.NowMonotonic().Unix())
	return a.appCtx, a.err
}

func (a *dbApp) doLoad(d *Database) (*config.AppContext, error) {
	data, err := d.Store.GetDatabaseSourceByAppID(a.appID)
	if err != nil {
		return nil, err
	}

	appFs, err := MakeAppFSFromDatabaseSource(data)
	if err != nil {
		return nil, err
	}
	resources := d.BaseResources.Overlay(appFs)

	appConfig, err := LoadConfig(resources)
	if err != nil {
		return nil, err
	}

	return &config.AppContext{
		AppFs:     appFs,
		Resources: resources,
		Config:    appConfig,
	}, nil
}
