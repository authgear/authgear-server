package webapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var whatsappWebhookLogger = slogutil.NewLogger("whatsapp-webhook")

type WhatsappCloudAPIWebhookWhatsappService interface {
	UpdateMessageStatus(ctx context.Context, messageID string, status whatsapp.WhatsappMessageStatus, errors []whatsapp.WhatsappStatusError) error
}

type WhatsappCloudAPIWebhookHandler struct {
	AppID           config.AppID
	WhatsappService WhatsappCloudAPIWebhookWhatsappService
	Credentials     *config.WhatsappCloudAPICredentials
	AppHostSuffixes config.AppHostSuffixes
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
	ID                    string                         `json:"id"`
	BizOpaqueCallbackData string                         `json:"biz_opaque_callback_data"`
	Status                string                         `json:"status"`
	Timestamp             string                         `json:"timestamp"`
	RecipientID           string                         `json:"recipient_id"`
	Errors                []whatsapp.WhatsappStatusError `json:"errors"`
}

func ConfigureWhatsappCloudAPIWebhookRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/whatsapp/webhook")
}

func (h *WhatsappCloudAPIWebhookHandler) handleVerifyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logger := whatsappWebhookLogger.GetLogger(ctx).With(
		slog.String("app_id", string(h.AppID)),
	)

	hubChallenge := r.URL.Query().Get("hub.challenge")
	hubVerifyToken := r.URL.Query().Get("hub.verify_token")

	if hubChallenge != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(hubChallenge))
		return
	}

	if subtle.ConstantTimeCompare([]byte(hubVerifyToken), []byte(h.Credentials.Webhook.VerifyToken)) != 1 {
		logger.Error(ctx, "invalid verify token received")
		http.Error(w, "invalid verify token", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WhatsappCloudAPIWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := whatsappWebhookLogger.GetLogger(ctx).With(
		slog.String("app_id", string(h.AppID)),
	)

	if h.Credentials == nil || h.Credentials.Webhook == nil {
		logger.Error(ctx, "whatsapp cloud api webhook credential is not configured")
		// Simply return 404 if webhook is not configured
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		h.handleVerifyRequest(ctx, w, r)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		logger.Error(ctx, "missing signature")
		http.Error(w, "missing signature", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(signature, "sha256=") {
		logger.Error(ctx, "invalid X-Hub-Signature-256 header format")
		http.Error(w, "invalid X-Hub-Signature-256 header format", http.StatusBadRequest)
		return
	}
	signature = strings.TrimPrefix(signature, "sha256=")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to read request body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mac := hmac.New(sha256.New, []byte(h.Credentials.Webhook.AppSecret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
		logger.Error(ctx, "invalid signature")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload whatsappWebhookPayload
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&payload)
	if err != nil {
		logger.WithError(err).Error(ctx, "invalid request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expectedBizOpaqueCallbackData := h.AppHostSuffixes.ToWhatsappCloudAPIBizOpaqueCallbackData()
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				if subtle.ConstantTimeCompare([]byte(change.Value.Metadata.PhoneNumberID), []byte(h.Credentials.PhoneNumberID)) != 1 {
					logger.Error(
						ctx,
						"phone number ID does not match configured phone number ID",
						slog.String("phone_number_id", change.Value.Metadata.PhoneNumberID),
					)
					continue
				}

				for _, status := range change.Value.Statuses {
					if expectedBizOpaqueCallbackData != "" {
						match := expectedBizOpaqueCallbackData == status.BizOpaqueCallbackData
						logger.Debug(ctx, "checking biz_opaque_callback_data", slog.Bool("match", match))
						if !match {
							continue
						}
					}

					err := h.WhatsappService.UpdateMessageStatus(
						ctx,
						status.ID,
						whatsapp.WhatsappMessageStatus(status.Status),
						status.Errors,
					)
					if err != nil {
						logger.WithError(err).Error(ctx, "Failed to update message status")
						continue
					}
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
