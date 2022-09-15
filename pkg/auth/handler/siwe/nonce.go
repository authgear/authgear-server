package siwe

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	featuresiwe "github.com/authgear/authgear-server/pkg/lib/feature/siwe"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureNonceRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/siwe/nonce")
}

type NonceHandlerSIWEService interface {
	CreateNewNonce() (*featuresiwe.Nonce, error)
}

type NonceHandlerLogger struct{ *log.Logger }

func NewNonceHandlerLogger(lf *log.Factory) NonceHandlerLogger {
	return NonceHandlerLogger{lf.New("handler-nonce")}
}

type NonceResponse struct {
	Nonce    string    `json:"nonce"`
	ExpireAt time.Time `json:"expire_at"`
}

type NonceHandlerJSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type NonceHandler struct {
	Logger NonceHandlerLogger
	SIWE   NonceHandlerSIWEService
	JSON   NonceHandlerJSONResponseWriter
}

func (h *NonceHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	nonce, err := h.SIWE.CreateNewNonce()
	if err != nil {
		h.Logger.WithError(err).Error("failed to create siwe nonce")
		http.Error(rw, "internal server error", 500)
		return
	}

	h.JSON.WriteResponse(rw, &api.Response{
		Result: &NonceResponse{
			Nonce:    nonce.Nonce,
			ExpireAt: nonce.ExpireAt,
		},
	})
}
