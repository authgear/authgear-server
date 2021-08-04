package healthz

import (
	"context"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type HandlerLogger struct{ *log.Logger }

func NewHandlerLogger(lf *log.Factory) HandlerLogger {
	return HandlerLogger{lf.New("healthz")}
}

type Handler struct {
	Context       context.Context
	AppDatabase   *appdb.Handle
	AppExecutor   *appdb.SQLExecutor
	AuditDatabase *auditdb.ReadHandle
	AuditExecutor *auditdb.ReadSQLExecutor
	Redis         *redis.Handle
	Logger        HandlerLogger
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := h.CheckHealth()
	if err != nil {
		h.Logger.WithError(err).Errorf("health check failed")
		http.Error(rw, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	body := []byte("OK")
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Content-Length", strconv.Itoa(len(body)))
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(body)
}

func (h *Handler) CheckHealth() (err error) {
	err = h.AppDatabase.ReadOnly(func() error {
		_, err := h.AppExecutor.QueryRowWith(sq.Select("42"))
		if err != nil {
			return err
		}
		h.Logger.Infof("app database connection healthz passed")
		return nil
	})
	if err != nil {
		return
	}

	if h.AuditDatabase != nil {
		err = h.AuditDatabase.ReadOnly(func() error {
			_, err := h.AuditExecutor.QueryRowWith(sq.Select("42"))
			if err != nil {
				return err
			}
			h.Logger.Infof("audit database connection healthz passed")
			return nil
		})
		if err != nil {
			return
		}
	}

	err = h.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Ping(h.Context).Result()
		if err != nil {
			return err
		}
		h.Logger.Infof("redis connection healthz passed")
		return nil
	})
	if err != nil {
		return
	}

	return
}
