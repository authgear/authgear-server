package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureWhatsappWATICallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/whatsapp/callback/wati")
}

type WhatsappWATICallbackHandler struct {
	WhatsappCodeProvider WhatsappCodeProvider
	Logger               WhatsappWATICallbackHandlerLogger
}

type WhatsappWATICallbackHandlerLogger struct{ *log.Logger }

func NewWhatsappWATICallbackHandlerLogger(lf *log.Factory) WhatsappWATICallbackHandlerLogger {
	return WhatsappWATICallbackHandlerLogger{lf.New("webapp-whatsapp-wati-callback-handler")}
}

func (h *WhatsappWATICallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("wati callback received")
}
