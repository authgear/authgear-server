package webapp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("whatsapp-webhook")

type WhatsappCloudAPIWebhookWhatsappService interface {
	UpdateMessageStatus(ctx context.Context, messageID string, status whatsapp.WhatsappMessageStatus) error
}

type WhatsappCloudAPIWebhookHandler struct {
	AppID           config.AppID
	WhatsappService WhatsappCloudAPIWebhookWhatsappService
	Credentials     *config.WhatsappCloudAPICredentials
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

func (h *WhatsappCloudAPIWebhookHandler) handleVerifyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	hubChallenge := r.URL.Query().Get("hub.challenge")
	hubVerifyToken := r.URL.Query().Get("hub.verify_token")

	if hubChallenge != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(hubChallenge))
		return
	}

	if subtle.ConstantTimeCompare([]byte(hubVerifyToken), []byte(h.Credentials.Webhook.VerifyToken)) != 1 {
		logger.GetLogger(ctx).With(
			slog.String("app_id", string(h.AppID)),
		).Error(ctx, "invalid verify token received")
		http.Error(w, "invalid verify token", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WhatsappCloudAPIWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.Credentials == nil || h.Credentials.Webhook == nil {
		logger.GetLogger(ctx).With(
			slog.String("app_id", string(h.AppID)),
		).Error(ctx, "whatsapp cloud api webhook credential is not configured")
		// Simply return 404 if webhook is not configured
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		h.handleVerifyRequest(ctx, w, r)
		return
	}

	var payload whatsappWebhookPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		// Log the error, because normally we won't receive any invalid request
		logger.GetLogger(ctx).With(
			slog.String("app_id", string(h.AppID)),
		).WithError(err).Error(ctx, "invalid request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				if subtle.ConstantTimeCompare([]byte(change.Value.Metadata.PhoneNumberID), []byte(h.Credentials.PhoneNumberID)) != 1 {
					logger.GetLogger(ctx).With(
						slog.String("app_id", string(h.AppID)),
						slog.String("phone_number_id", change.Value.Metadata.PhoneNumberID),
					).Error(ctx, "phone number ID does not match configured phone number ID")
					continue
				}

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
