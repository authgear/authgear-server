package configsource

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/afero"
	// We need "sigs.k8s.io/yaml" package instead of other yaml serializer,
	// because "gopkg.in/yaml.v3" add `null`s for null pointers which break some validations.
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const PGChannelConfigSourceChange = "config_source_change"
const PGChannelDomainChange = "domain_change"
const PGChannelPlanChange = "plan_change"

type DatabaseSource struct {
	ID        string
	AppID     string
	Data      map[string][]byte
	PlanName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DatabaseLogger struct{ *log.Logger }

func NewDatabaseLogger(lf *log.Factory) DatabaseLogger {
	return DatabaseLogger{lf.New("configsource-database")}
}

type ResolveAppIDType string

func NewResolveAppIDTypeDomain() ResolveAppIDType {
	return ResolveAppIDTypeDomain
}

func NewResolveAppIDTypePath() ResolveAppIDType {
	return ResolveAppIDTypePath
}

const (
	ResolveAppIDTypeDomain ResolveAppIDType = "domain"
	ResolveAppIDTypePath   ResolveAppIDType = "path"
)

type DatabaseHandleFactory func() *globaldb.Handle
type ConfigSourceStoreFactory func(handle *globaldb.Handle) *Store
type PlanStoreFactory plan.StoreFactory

func NewDatabaseHandleFactory(
	pool *db.Pool,
	credentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
	cfg *config.DatabaseEnvironmentConfig,
	lf *log.Factory,
) DatabaseHandleFactory {
	factory := func() *globaldb.Handle {
		return globaldb.NewHandle(
			pool,
			credentials,
			cfg,
			lf,
		)
	}
	return factory
}

func NewConfigSourceStoreStoreFactory(
	sqlbuilder *globaldb.SQLBuilder,
) ConfigSourceStoreFactory {
	factory := func(handle *globaldb.Handle) *Store {
		sqlExecutor := globaldb.NewSQLExecutor(handle)
		return &Store{
			SQLBuilder:  sqlbuilder,
			SQLExecutor: sqlExecutor,
		}
	}
	return factory
}

func NewPlanStoreStoreFactory(
	sqlbuilder *globaldb.SQLBuilder,
) PlanStoreFactory {
	factory := PlanStoreFactory(plan.NewStoreFactory(sqlbuilder))
	return factory
}

type Database struct {
	Logger                   DatabaseLogger
	BaseResources            *resource.Manager
	TrustProxy               config.TrustProxy
	Config                   *Config
	Clock                    clock.Clock
	ConfigSourceStoreFactory ConfigSourceStoreFactory
	PlanStoreFactory         PlanStoreFactory
	DatabaseHandleFactory    DatabaseHandleFactory
	DatabaseCredentials      *config.GlobalDatabaseCredentialsEnvironmentConfig
	DatabaseConfig           *config.DatabaseEnvironmentConfig

	ResolveAppIDType ResolveAppIDType

	done     chan<- struct{} `wire:"-"`
	listener *db.PQListener  `wire:"-"`

	hostMap *sync.Map `wire:"-"`
	appMap  *sync.Map `wire:"-"`
}

var _ ContextResolver = &Database{}

func (d *Database) Open(ctx context.Context) error {
	d.hostMap = &sync.Map{}
	d.appMap = &sync.Map{}

	done := make(chan struct{})
	d.done = done

	d.listener = &db.PQListener{
		DatabaseURL: d.DatabaseCredentials.DatabaseURL,
	}
	go d.listener.Listen([]string{
		PGChannelConfigSourceChange,
		PGChannelDomainChange,
		PGChannelPlanChange,
	}, done, func(channel string, extra string) {
		switch channel {
		case PGChannelConfigSourceChange:
			d.invalidateApp(extra)
		case PGChannelDomainChange:
			d.invalidateHost(extra)
			d.invalidateAppByDomain(ctx, extra)
		case PGChannelPlanChange:
			d.invalidateAllApp()
		default:
			// unknown notification channel, just skip it
			d.Logger.WithField("channel", channel).Info("unknown notification channel")
		}
	}, func(e error) {
		d.Logger.WithError(e).Error("error on listening pgsql")
	})
	go d.cleanupCache(done)

	return nil
}

func (d *Database) Close() error {
	close(d.done)
	return nil
}

func (d *Database) ResolveAppID(ctx context.Context, r *http.Request) (string, error) {
	switch d.ResolveAppIDType {
	case ResolveAppIDTypeDomain:
		return d.resolveAppIDByDomain(ctx, r)
	case ResolveAppIDTypePath:
		return d.resolveAppIDByPath(ctx, r)
	default:
		panic("invalid resolve app id type")
	}
}

func (d *Database) resolveAppIDByPath(ctx context.Context, r *http.Request) (string, error) {
	appid := httproute.GetParam(r, "appid")
	if appid == "" {
		return "", ErrAppNotFound
	}
	// Try to resolve app to ensure the app exist
	_, _, err := d.ResolveContext(ctx, appid)
	if err != nil {
		return "", err
	}

	return appid, nil
}

func (d *Database) resolveAppIDByDomain(ctx context.Context, r *http.Request) (string, error) {
	host := httputil.GetHost(r, bool(d.TrustProxy))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	appIDData, ok := d.hostMap.Load(host)
	if ok {
		return appIDData.(string), nil
	}

	var appID string
	dbHandle := d.DatabaseHandleFactory()
	store := d.ConfigSourceStoreFactory(dbHandle)
	err := dbHandle.WithTx(ctx, func(ctx context.Context) error {
		d.Logger.WithField("host", host).Debug("resolve appid from db")
		aid, err := store.GetAppIDByDomain(ctx, host)
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

func (d *Database) ResolveContext(ctx context.Context, appID string) (context.Context, *config.AppContext, error) {
	value, _ := d.appMap.LoadOrStore(appID, &dbApp{
		appID: appID,
		load:  &sync.Once{},
	})
	app := value.(*dbApp)
	appCtx, err := app.Load(ctx, d)
	if err != nil {
		return nil, nil, err
	}
	ctx = config.WithAppContext(ctx, appCtx)

	return ctx, appCtx, nil
}

func (d *Database) ReloadApp(ctx context.Context, appID string) {
	d.invalidateApp(appID)
}

func (d *Database) CreateDatabaseSource(ctx context.Context, appID string, resources map[string][]byte, planName string) error {
	dbHandle := d.DatabaseHandleFactory()
	store := d.ConfigSourceStoreFactory(dbHandle)
	return dbHandle.WithTx(ctx, func(ctx context.Context) error {
		_, err := store.GetDatabaseSourceByAppID(ctx, appID)
		if err != nil && !errors.Is(err, ErrAppNotFound) {
			return err
		} else if err == nil {
			return ErrDuplicatedAppID
		}

		dbData := make(map[string][]byte)
		for path, data := range resources {
			dbData[filepathutil.EscapePath(path)] = data
		}

		dbSource := &DatabaseSource{
			ID:        uuid.New(),
			AppID:     appID,
			Data:      dbData,
			PlanName:  planName,
			CreatedAt: d.Clock.NowUTC(),
			UpdatedAt: d.Clock.NowUTC(),
		}
		return store.CreateDatabaseSource(ctx, dbSource)
	})
}

func (d *Database) UpdateDatabaseSource(ctx context.Context, appID string, updates []*resource.ResourceFile) error {
	dbHandle := d.DatabaseHandleFactory()
	store := d.ConfigSourceStoreFactory(dbHandle)
	return dbHandle.WithTx(ctx, func(ctx context.Context) error {
		dbs, err := store.GetDatabaseSourceByAppID(ctx, appID)
		if err != nil {
			return err
		}

		updated := false
		for _, u := range updates {
			key := filepathutil.EscapePath(u.Location.Path)
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
			err = store.UpdateDatabaseSource(ctx, dbs)
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

func (d *Database) invalidateAllApp() {
	d.appMap.Range(func(key, value any) bool {
		d.appMap.Delete(key)
		return true
	})
	d.Logger.Info("invalidated all cached config")
}

func (d *Database) invalidateAppByDomain(ctx context.Context, domain string) {
	dbHandle := d.DatabaseHandleFactory()
	store := d.ConfigSourceStoreFactory(dbHandle)
	err := dbHandle.WithTx(ctx, func(ctx context.Context) error {
		aid, err := store.GetAppIDByDomain(ctx, domain)
		if err != nil {
			return err
		}
		d.invalidateApp(aid)
		return nil
	})
	if err != nil {
		d.Logger.WithError(err).
			WithField("domain", domain).
			Errorln("failed to invalidate app cache by domain")
	}
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

func newMemMapFs() afero.Fs {
	// Construct a FS that treats `a` and `/a` the same.
	// The template is loaded by a file URI which is always an absoluted path.
	return afero.NewBasePathFs(afero.NewMemMapFs(), "/")
}

func MakeAppFSFromDatabaseSource(s *DatabaseSource) (resource.Fs, error) {
	appFs := newMemMapFs()
	create := func(name string, data []byte) {
		file, _ := appFs.Create(name)
		_, _ = file.Write(data)
	}

	for key, data := range s.Data {
		path, err := filepathutil.UnescapePath(key)
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

func MakePlanFSFromPlan(p *plan.Plan) (resource.Fs, error) {
	planFs := newMemMapFs()
	if p != nil {
		file, _ := planFs.Create(AuthgearFeatureYAML)
		data, err := yaml.Marshal(p.RawFeatureConfig)
		if err != nil {
			return nil, err
		}
		_, _ = file.Write(data)
	}
	return &resource.LeveledAferoFs{
		Fs:      planFs,
		FsLevel: resource.FsLevelPlan,
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

func (a *dbApp) Load(ctx context.Context, d *Database) (*config.AppContext, error) {
	if a.load != nil {
		a.load.Do(func() {
			a.appCtx, a.err = a.doLoad(ctx, d)
		})
	}
	atomic.StoreInt64(&a.lastUsedAt, d.Clock.NowMonotonic().Unix())
	return a.appCtx, a.err
}

func (a *dbApp) doLoad(ctx context.Context, d *Database) (*config.AppContext, error) {
	var appCtx *config.AppContext
	dbHandle := d.DatabaseHandleFactory()
	store := d.ConfigSourceStoreFactory(dbHandle)
	planStore := d.PlanStoreFactory(dbHandle)
	err := dbHandle.WithTx(ctx, func(ctx context.Context) error {
		d.Logger.WithField("app_id", a.appID).Info("load app config from db")
		data, err := store.GetDatabaseSourceByAppID(ctx, a.appID)
		if err != nil {
			return err
		}

		p, err := planStore.GetPlan(ctx, data.PlanName)
		if err != nil && !errors.Is(err, plan.ErrPlanNotFound) {
			return err
		}

		planFs, err := MakePlanFSFromPlan(p)
		if err != nil {
			return err
		}

		appFs, err := MakeAppFSFromDatabaseSource(data)
		if err != nil {
			return err
		}
		resources := d.BaseResources.Overlay(planFs)
		resources = resources.Overlay(appFs)

		appConfig, err := LoadConfig(ctx, resources)
		if err != nil {
			return err
		}

		domains, err := store.GetDomainsByAppID(ctx, a.appID)
		if err != nil {
			return err
		}

		appCtx = &config.AppContext{
			AppFs:     appFs,
			Resources: resources,
			Config:    appConfig,
			PlanName:  data.PlanName,
			Domains:   config.AppDomains(domains),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return appCtx, nil
}
