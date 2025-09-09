package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureWhatsappCloudAPIWebhookRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/whatsapp/webhook")
}

type WhatsappCloudAPIWebhookHandler struct {
}

func (h *WhatsappCloudAPIWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement webhook handling logic
	w.WriteHeader(http.StatusOK)
}
