package healthz

import (
	"context"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type HandlerLogger struct{ *log.Logger }

func NewHandlerLogger(lf *log.Factory) HandlerLogger {
	return HandlerLogger{lf.New("healthz")}
}

type Handler struct {
	GlobalDatabase *globaldb.Handle
	GlobalExecutor *globaldb.SQLExecutor
	Logger         HandlerLogger
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := h.CheckHealth(r.Context())
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

func (h *Handler) CheckHealth(ctx context.Context) (err error) {
	err = h.GlobalDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		var fortyTwo int
		row, err := h.GlobalExecutor.QueryRowWith(ctx, sq.Select("42"))
		if err != nil {
			return err
		}
		err = row.Scan(&fortyTwo)
		if err != nil {
			return err
		}

		h.Logger.Debugf("global database connection healthz passed")
		return nil
	})
	if err != nil {
		return
	}

	return
}
