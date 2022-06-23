package transport

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/service"
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
	CreateSubscriptionIfNotExists(stripeCheckoutSessionID string) error
}

type SubscriptionService interface {
	UpdateSubscriptionCheckoutStatusAndCustomerID(appID string, stripCheckoutSessionID string, status model.SubscriptionCheckoutStatus, customerID string) error
	UpdateSubscriptionCheckoutStatusByCustomerID(appID string, customerID string, status model.SubscriptionCheckoutStatus) error
	CreateSubscription(appID string, stripeSubscriptionID string, stripeCustomerID string) (*model.Subscription, error)
}

type StripeWebhookHandler struct {
	StripeService StripeService
	Logger        StripeWebhookLogger
	Subscriptions SubscriptionService
	Database      *globaldb.Handle
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
	case libstripe.EventTypeCustomerSubscriptionCreated:
		err = h.handleCustomerSubscriptionEvent(
			event.(*libstripe.CustomerSubscriptionCreatedEvent).CustomerSubscriptionEvent,
		)
	case libstripe.EventTypeCustomerSubscriptionUpdated:
		err = h.handleCustomerSubscriptionEvent(
			event.(*libstripe.CustomerSubscriptionUpdatedEvent).CustomerSubscriptionEvent,
		)
	}
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompletedEvent(event *libstripe.CheckoutSessionCompletedEvent) error {
	// Update _portal_subscription_checkout set state=completed, stripe_customer_id
	err := h.Database.WithTx(func() error {
		return h.Subscriptions.UpdateSubscriptionCheckoutStatusAndCustomerID(
			event.AppID,
			event.StripeCheckoutSessionID,
			model.SubscriptionCheckoutStatusCompleted,
			event.StripeCustomerID,
		)
	})
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionCheckoutNotFound) {
			// The checkout is not found or the checkout is already subscribed
			// Tolerate it.
			h.Logger.
				WithField("app_id", event.AppID).
				WithField("stripe_checkout_session_id", event.StripeCheckoutSessionID).
				Info("the subscription checkout does not exists or the status is subscribed already")
			return nil
		}
		return err
	}

	// Check and create stripe subscription
	err = h.Database.WithTx(func() error {
		return h.StripeService.CreateSubscriptionIfNotExists(event.StripeCheckoutSessionID)
	})
	if err != nil {
		if errors.Is(err, libstripe.ErrCustomerAlreadySubscribed) {
			// The customer has subscriptions already
			// Tolerate it
			h.Logger.
				WithField("app_id", event.AppID).
				WithField("stripe_checkout_session_id", event.StripeCheckoutSessionID).
				Info("customer already subscribed")
			return nil
		}
		return err
	}

	return nil
}

func (h *StripeWebhookHandler) handleCustomerSubscriptionEvent(event *libstripe.CustomerSubscriptionEvent) error {
	if !event.IsSubscriptionActive() {
		// If the subscription event is about status that are not active
		// Ignore it
		h.Logger.
			WithField("stripe_subscription_id", event.StripeSubscriptionID).
			WithField("stripe_subscription_status", event.StripeSubscriptionStatus).
			Info("unhandled subscription status")
		return nil
	}

	err := h.Database.WithTx(func() error {
		// Mark checkout session as subscribed
		err := h.Subscriptions.UpdateSubscriptionCheckoutStatusByCustomerID(
			event.AppID,
			event.StripeCustomerID,
			model.SubscriptionCheckoutStatusSubscribed,
		)
		if err != nil {
			if errors.Is(err, service.ErrSubscriptionCheckoutNotFound) {
				// The checkout is not found or the checkout is already subscribed
				// Tolerate it.
				h.Logger.
					WithField("app_id", event.AppID).
					WithField("stripe_subscription_id", event.StripeSubscriptionID).
					Info("the subscription checkout does not exists or the status is subscribed already")
				return nil
			}
			return err
		}

		// Insert _portal_subscription
		_, err = h.Subscriptions.CreateSubscription(event.AppID, event.StripeSubscriptionID, event.StripeCustomerID)
		if err != nil {
			return err
		}

		// FIXME(billing): update app plan

		return nil
	})
	return err
}
