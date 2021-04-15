package configsource

import (
	"net"
	"net/http"
	"time"

	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type DatabaseSource struct {
	ID        string
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

	return "", nil
}

func (d *Database) ResolveContext(appID string) (*config.AppContext, error) {
	return nil, nil
}

func (d *Database) ReloadApp(appID string) {
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
