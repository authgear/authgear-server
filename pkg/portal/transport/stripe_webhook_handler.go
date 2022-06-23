package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureStripeWebhookRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").WithPathPattern("/api/subscription/webhook/stripe")
}

type StripeWebhookLogger struct{ *log.Logger }

func NewStripeWebhookLogger(lf *log.Factory) StripeWebhookLogger {
	return StripeWebhookLogger{lf.New("stripe-webhook")}
}

type StripeService interface {
	ConstructEvent(r *http.Request) (libstripe.Event, error)
}

type SubscriptionService interface {
}

type StripeWebhookHandler struct {
	StripeService StripeService
	Logger        StripeWebhookLogger
	Subscriptions SubscriptionService
}

func (h *StripeWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			h.Logger.WithError(err).Errorf("failed to handle stripe webhook")
			http.Error(w, "failed to handle stripe webhook", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}()

	event, err := h.StripeService.ConstructEvent(r)
	if err != nil {
		if errors.Is(err, libstripe.ErrUnknownEvent) {
			// It is common to receive unknown event
			// e.g. create objects via stripe portal doesn't have the metadata
			//      event type that are not handled by the server
			// gracefully ignore them
			err = nil
			w.WriteHeader(http.StatusOK)
			return
		}
		return
	}

	h.Logger.
		WithField("event_type", event.EventType()).
		Info("stripe webhook event received")

	switch event.EventType() {
	case libstripe.EventTypeCheckoutSessionCompleted:
		err = h.handleCheckoutSessionCompletedEvent(event.(*libstripe.CheckoutSessionCompletedEvent))
	}
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompletedEvent(event *libstripe.CheckoutSessionCompletedEvent) error {
	fmt.Println("handleCheckoutSessionCompletedEvent", event)
	return nil
}
