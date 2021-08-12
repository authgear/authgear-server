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
	Context        context.Context
	GlobalDatabase *globaldb.Handle
	GlobalExecutor *globaldb.SQLExecutor
	Logger         HandlerLogger
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
	err = h.GlobalDatabase.ReadOnly(func() error {
		_, err := h.GlobalExecutor.QueryRowWith(sq.Select("42"))
		if err != nil {
			return err
		}
		h.Logger.Infof("global database connection healthz passed")
		return nil
	})
	if err != nil {
		return
	}

	return
}
