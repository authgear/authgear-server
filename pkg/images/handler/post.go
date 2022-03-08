package handler

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigurePostRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST").
		WithPathPattern("/_images/:appid/:objectid")
}

type PostHandlerLogger struct{ *log.Logger }

func NewPostHandlerLogger(lf *log.Factory) PostHandlerLogger {
	return PostHandlerLogger{lf.New("post-handler")}
}

type PostHandler struct {
	Logger PostHandlerLogger
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO(images): implement post image endpint
	http.Error(w, "not implemented", http.StatusBadRequest)
}
