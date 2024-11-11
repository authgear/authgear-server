package transport

import (
	"context"
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
	CreateSubscriptionIfNotExists(ctx context.Context, stripeCheckoutSessionID string, subscriptionPlans []*model.SubscriptionPlan) error
	FetchSubscriptionPlans(ctx context.Context) (subscriptionPlans []*model.SubscriptionPlan, err error)
}

type SubscriptionService interface {
	GetSubscription0(ctx context.Context, appID string) (*model.Subscription, error)
	MarkCheckoutSubscribed0(ctx context.Context, appID string, customerID string) error
	MarkCheckoutCancelled0(ctx context.Context, appID string, customerID string) error
	UpsertSubscription0(ctx context.Context, appID string, stripeSubscriptionID string, stripeCustomerID string) (*model.Subscription, error)
	ArchiveSubscription0(ctx context.Context, sub *model.Subscription) error
	UpdateAppPlan0(ctx context.Context, appID string, planName string) error
	UpdateAppPlanToDefault0(ctx context.Context, appID string) error

	MarkCheckoutCompleted(ctx context.Context, appID string, stripCheckoutSessionID string, customerID string) error
	MarkCheckoutExpired(ctx context.Context, appID string, customerID string) error
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
		if errors.Is(err, libstripe.ErrUnknownEvent) {
			// It is common to receive unknown event
			// e.g. create objects via stripe portal doesn't have the metadata
			//      event type that are not handled by the server
			// gracefully ignore them
			w.WriteHeader(http.StatusOK)
			return
		}
		if err != nil {
			h.Logger.WithError(err).Errorf("failed to handle stripe webhook")
			http.Error(w, "failed to handle stripe webhook", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}()

	event, err := h.StripeService.ConstructEvent(r)
	if err != nil {
		return
	}

	h.Logger.
		WithField("event_type", event.EventType()).
		Info("stripe webhook event received")

	switch event.EventType() {
	case libstripe.EventTypeCheckoutSessionCompleted:
		err = h.handleCheckoutSessionCompletedEvent(r.Context(), event.(*libstripe.CheckoutSessionCompletedEvent))
	case libstripe.EventTypeCustomerSubscriptionCreated:
		err = h.handleCustomerSubscriptionEvent(
			r.Context(),
			event.(*libstripe.CustomerSubscriptionCreatedEvent).CustomerSubscriptionEvent,
		)
	case libstripe.EventTypeCustomerSubscriptionUpdated:
		// FIXME(subscription): handle update subscription
		err = h.handleCustomerSubscriptionEvent(
			r.Context(),
			event.(*libstripe.CustomerSubscriptionUpdatedEvent).CustomerSubscriptionEvent,
		)
	case libstripe.EventTypeCustomerSubscriptionDeleted:
		err = h.handleCustomerSubscriptionDeletedEvent(
			r.Context(),
			event.(*libstripe.CustomerSubscriptionDeletedEvent).CustomerSubscriptionEvent,
		)
	}
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompletedEvent(ctx context.Context, event *libstripe.CheckoutSessionCompletedEvent) error {
	// Update _portal_subscription_checkout set state=completed, stripe_customer_id
	err := h.Subscriptions.MarkCheckoutCompleted(
		ctx,
		event.AppID,
		event.StripeCheckoutSessionID,
		event.StripeCustomerID,
	)
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
	var subscriptionPlan []*model.SubscriptionPlan
	err = h.Database.ReadOnly(ctx, func(ctx context.Context) error {
		subscriptionPlan, err = h.StripeService.FetchSubscriptionPlans(ctx)
		if err != nil {
			return err
		}
		return nil
	})

	err = h.StripeService.CreateSubscriptionIfNotExists(ctx, event.StripeCheckoutSessionID, subscriptionPlan)
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
		if errors.Is(err, libstripe.ErrAppAlreadySubscribed) {
			// The app has stripe subscription already
			// Tolerate it
			h.Logger.
				WithField("app_id", event.AppID).
				WithField("stripe_checkout_session_id", event.StripeCheckoutSessionID).
				Warn("app already has stripe subscription")
			return nil
		}
		return err
	}

	return nil
}

func (h *StripeWebhookHandler) handleCustomerSubscriptionEvent(ctx context.Context, event *libstripe.CustomerSubscriptionEvent) error {
	// Here is a complete list of subscription status and our corresponding action.
	// incomplete -> ignore
	// incomplete_expired -> set checkout to cancelled.
	// trialing -> ignore
	// active -> set checkout to subscribed
	// past_due -> ignore
	// canceled -> ignore
	// unpaid -> ignore

	if event.IsSubscriptionActive() {
		return h.handleActiveSubscriptionEvent(ctx, event)
	}

	if event.IsSubscriptionIncompleteExpired() {
		return h.handleIncompleteExpiredSubscriptionEvent(ctx, event)
	}

	h.Logger.
		WithField("stripe_subscription_id", event.StripeSubscriptionID).
		WithField("stripe_subscription_status", event.StripeSubscriptionStatus).
		Info("unhandled subscription status")
	return nil
}

func (h *StripeWebhookHandler) handleIncompleteExpiredSubscriptionEvent(ctx context.Context, event *libstripe.CustomerSubscriptionEvent) error {
	err := h.Subscriptions.MarkCheckoutExpired(
		ctx,
		event.AppID,
		event.StripeCustomerID,
	)
	if err != nil {
		if !errors.Is(err, service.ErrSubscriptionCheckoutNotFound) {
			return err
		}
		// The checkout session doesn't exist
		// It may happen if the subscription is created via Stripe portal
		// Tolerate it.
		h.Logger.
			WithField("app_id", event.AppID).
			WithField("stripe_subscription_id", event.StripeSubscriptionID).
			Info("the subscription checkout does not exist for incomplete_expired")
		return nil
	}
	return nil
}

func (h *StripeWebhookHandler) handleActiveSubscriptionEvent(ctx context.Context, event *libstripe.CustomerSubscriptionEvent) error {
	err := h.Database.WithTx(ctx, func(ctx context.Context) error {
		// Mark checkout session as subscribed
		err := h.Subscriptions.MarkCheckoutSubscribed0(
			ctx,
			event.AppID,
			event.StripeCustomerID,
		)
		if err != nil {
			if !errors.Is(err, service.ErrSubscriptionCheckoutNotFound) {
				return err
			}
			// The checkout is not found or the checkout is already subscribed
			// Tolerate it.
			h.Logger.
				WithField("app_id", event.AppID).
				WithField("stripe_subscription_id", event.StripeSubscriptionID).
				Info("the subscription checkout does not exists or the status is subscribed already")
			// Fallthrough here so subscription will be upserted.
		}

		// Upsert _portal_subscription
		_, err = h.Subscriptions.UpsertSubscription0(ctx, event.AppID, event.StripeSubscriptionID, event.StripeCustomerID)
		if err != nil {
			return err
		}

		// Update app plan
		err = h.Subscriptions.UpdateAppPlan0(ctx, event.AppID, event.PlanName)
		if err != nil {
			return err
		}
		h.Logger.
			WithField("app_id", event.AppID).
			WithField("plan_name", event.PlanName).
			Info("updated app plan")

		return nil
	})
	return err
}

func (h *StripeWebhookHandler) handleCustomerSubscriptionDeletedEvent(ctx context.Context, event *libstripe.CustomerSubscriptionEvent) error {
	if !event.IsSubscriptionCanceled() {
		// The status should be cancelled in the `customer.subscription.deleted` event
		// In case it is not, log it as warning and ignore it
		h.Logger.
			WithField("stripe_subscription_id", event.StripeSubscriptionID).
			WithField("stripe_subscription_status", event.StripeSubscriptionStatus).
			Warn("unexpected subscription status, it should be cancelled")
		return nil
	}

	err := h.Database.WithTx(ctx, func(ctx context.Context) error {
		// Mark checkout session as cancelled
		err := h.Subscriptions.MarkCheckoutCancelled0(
			ctx,
			event.AppID,
			event.StripeCustomerID,
		)
		if err != nil {
			if !errors.Is(err, service.ErrSubscriptionCheckoutNotFound) {
				return err
			}
			// The checkout session doesn't exist
			// It may happen if the subscription is created via Stripe portal
			// Tolerate it.
			h.Logger.
				WithField("app_id", event.AppID).
				WithField("stripe_subscription_id", event.StripeSubscriptionID).
				Info("the subscription checkout does not exist for cancellation")
		}

		sub, err := h.Subscriptions.GetSubscription0(ctx, event.AppID)
		if err != nil {
			if errors.Is(err, service.ErrSubscriptionNotFound) {
				// Subscription doesn't exist in the db.
				// Ignore the event.
				h.Logger.
					WithField("app_id", event.AppID).
					WithField("stripe_subscription_id", event.StripeSubscriptionID).
					Warn("the subscription does not exist for cancellation")
				return nil
			}
			return err
		}

		if sub.StripeSubscriptionID != event.StripeSubscriptionID {
			// The cancelled subscription id doesn't match the one in the db.
			// It may happen if the subscription is managed in Stripe portal manually.
			// Ignore the event.
			h.Logger.
				WithField("app_id", event.AppID).
				WithField("stripe_subscription_id", event.StripeSubscriptionID).
				Warn("the subscription id doesn't match the one in the db for cancellation")
			return nil
		}

		err = h.Subscriptions.ArchiveSubscription0(ctx, sub)
		if err != nil {
			return err
		}

		err = h.Subscriptions.UpdateAppPlanToDefault0(ctx, event.AppID)
		if err != nil {
			return err
		}
		h.Logger.
			WithField("app_id", event.AppID).
			Info("cancelled app plan")

		return nil
	})
	return err
}
