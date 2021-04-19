package configsource

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const PGChannelConfigSourceChange = "config_source_change"
const PGChannelDomainChange = "domain_change"

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
	Logger         DatabaseLogger
	BaseResources  *resource.Manager
	TrustProxy     config.TrustProxy
	Config         *Config
	Clock          clock.Clock
	Store          *Store
	Database       *globaldb.Handle
	DatabaseConfig *config.DatabaseEnvironmentConfig

	done     chan<- struct{} `wire:"-"`
	listener *db.PQListener  `wire:"-"`

	hostMap *sync.Map `wire:"-"`
	appMap  *sync.Map `wire:"-"`
}

func (d *Database) Open() error {
	d.hostMap = &sync.Map{}
	d.appMap = &sync.Map{}

	done := make(chan struct{})
	d.done = done

	d.listener = &db.PQListener{
		DatabaseURL: d.DatabaseConfig.DatabaseURL,
	}
	go d.listener.Listen([]string{
		PGChannelConfigSourceChange,
		PGChannelDomainChange,
	}, done, func(channel string, extra string) {
		switch channel {
		case PGChannelConfigSourceChange:
			d.invalidateApp(extra)
		case PGChannelDomainChange:
			d.invalidateHost(extra)
		default:
			panic(fmt.Sprintf("db_config: unknown notification channel: %s", channel))
		}
	}, func(e error) {
		panic(fmt.Sprintf("db_config: error on listening pgsql: %s", e))
	})
	go d.cleanupCache(done)

	return nil
}

func (d *Database) Close() error {
	close(d.done)
	return nil
}

func (d *Database) ResolveAppID(r *http.Request) (string, error) {
	host := httputil.GetHost(r, bool(d.TrustProxy))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	appIDData, ok := d.hostMap.Load(host)
	if ok {
		return appIDData.(string), nil
	}

	var appID string
	err := d.Database.WithTx(func() error {
		d.Logger.WithField("host", host).Debug("resolve appid from db")
		aid, err := d.Store.GetAppIDByDomain(host)
		if err != nil {
			return err
		}
		d.hostMap.Store(host, aid)
		appID = aid
		return nil
	})

	if err != nil {
		return "", err
	}
	return appID, nil
}

func (d *Database) ResolveContext(appID string) (*config.AppContext, error) {
	value, _ := d.appMap.LoadOrStore(appID, &dbApp{
		appID: appID,
		load:  &sync.Once{},
	})
	app := value.(*dbApp)
	return app.Load(d)
}

func (d *Database) ReloadApp(appID string) {
	d.invalidateApp(appID)
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

func (d *Database) invalidateHost(domain string) {
	d.hostMap.Delete(domain)
	d.Logger.WithField("domain", domain).Info("invalidated cached host")
}

func (d *Database) invalidateApp(appID string) {
	d.appMap.Delete(appID)
	d.Logger.WithField("app_id", appID).Info("invalidated cached config")
}

func (d *Database) cleanupCache(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return

		case <-time.After(time.Minute):
			now := d.Clock.NowMonotonic().Unix()
			numDel := 0
			d.appMap.Range(func(key, value interface{}) bool {
				app := value.(*dbApp)
				if atomic.LoadInt64(&app.lastUsedAt) < now-60 {
					d.appMap.Delete(key)
					numDel++
				}
				return true
			})
			if numDel > 0 {
				d.Logger.WithField("deleted", numDel).Info("cleaned cached app configs")
			}
		}
	}
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
	load       *sync.Once
	Loaded     bool
}

func (a *dbApp) Load(d *Database) (*config.AppContext, error) {
	if a.load != nil {
		a.load.Do(func() {
			a.appCtx, a.err = a.doLoad(d)
		})
	}
	atomic.StoreInt64(&a.lastUsedAt, d.Clock.NowMonotonic().Unix())
	return a.appCtx, a.err
}

func (a *dbApp) doLoad(d *Database) (*config.AppContext, error) {
	var appCtx *config.AppContext
	err := d.Database.WithTx(func() error {
		d.Logger.WithField("app_id", a.appID).Info("load app config from db")
		data, err := d.Store.GetDatabaseSourceByAppID(a.appID)
		if err != nil {
			return err
		}

		appFs, err := MakeAppFSFromDatabaseSource(data)
		if err != nil {
			return err
		}
		resources := d.BaseResources.Overlay(appFs)

		appConfig, err := LoadConfig(resources)
		if err != nil {
			return err
		}

		appCtx = &config.AppContext{
			AppFs:     appFs,
			Resources: resources,
			Config:    appConfig,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return appCtx, nil
}
