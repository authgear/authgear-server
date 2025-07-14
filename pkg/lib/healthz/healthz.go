package healthz

import (
	"context"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var HealthzLogger = slogutil.NewLogger("healthz")

type Handler struct {
	GlobalDatabase *globaldb.Handle
	GlobalExecutor *globaldb.SQLExecutor
	GlobalRedis    *globalredis.Handle
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := HealthzLogger.GetLogger(ctx)
	err := h.CheckHealth(ctx)
	if err != nil {
		logger.WithError(err).Error(ctx, "health check failed")
		http.Error(rw, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	body := []byte("OK")
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Content-Length", strconv.Itoa(len(body)))
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(body)
}

func (h *Handler) CheckHealth(ctx context.Context) (err error) {
	err = h.GlobalDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		logger := HealthzLogger.GetLogger(ctx)

		var fortyTwo int
		row, err := h.GlobalExecutor.QueryRowWith(ctx, sq.Select("42"))
		if err != nil {
			return err
		}
		err = row.Scan(&fortyTwo)
		if err != nil {
			return err
		}

		logger.Debug(ctx, "global database connection healthz passed")
		return nil
	})
	if err != nil {
		return
	}

	err = h.GlobalRedis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		logger := HealthzLogger.GetLogger(ctx)

		_, err := conn.Ping(ctx).Result()
		if err != nil {
			return err
		}
		logger.Debug(ctx, "global redis connection healthz passed")
		return nil
	})
	if err != nil {
		return
	}

	return
}
