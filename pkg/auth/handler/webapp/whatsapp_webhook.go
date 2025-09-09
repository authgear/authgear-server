package webapp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("whatsapp-webhook")

type WhatsappCloudAPIWebhookWhatsappService interface {
	UpdateMessageStatus(ctx context.Context, messageID string, status whatsapp.WhatsappMessageStatus) error
}

type WhatsappCloudAPIWebhookHandler struct {
	WhatsappService WhatsappCloudAPIWebhookWhatsappService
}

type whatsappWebhookPayload struct {
	Entry []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Statuses []whatsappStatus `json:"statuses"`
			} `json:"value"`
			Field string `json:"field"`
		} `json:"changes"`
	} `json:"entry"`
}

type whatsappStatus struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	RecipientID string `json:"recipient_id"`
	Errors      []struct {
		Code    int    `json:"code"`
		Title   string `json:"title"`
		Message string `json:"message"`
	} `json:"errors"`
}

func ConfigureWhatsappCloudAPIWebhookRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/whatsapp/webhook")
}

func (h *WhatsappCloudAPIWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method == "GET" {
		// TODO: Handle webhook verification
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload whatsappWebhookPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				for _, status := range change.Value.Statuses {
					err := h.WhatsappService.UpdateMessageStatus(
						ctx,
						status.ID,
						whatsapp.WhatsappMessageStatus(status.Status),
					)
					if err != nil {
						logger.GetLogger(ctx).WithError(err).Error(ctx, "Failed to update message status")
						continue
					}
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
