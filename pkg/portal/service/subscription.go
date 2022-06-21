package service

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type SubscriptionService struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *SubscriptionService) CreateSubscription(stripeSubscription *libstripe.Subscription) (*model.Subscription, error) {
	subscription := &model.Subscription{
		ID:                      uuid.New(),
		AppID:                   stripeSubscription.AppID,
		StripeCheckoutSessionID: stripeSubscription.StripeCheckoutSessionID,
		StripeCustomerID:        stripeSubscription.StripeCustomerID,
		StripeSubscriptionID:    stripeSubscription.StripeSubscriptionID,
	}

	if err := s.createSubscription(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

func (s *SubscriptionService) createSubscription(sub *model.Subscription) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_subscription")).
		Columns(
			"id",
			"app_id",
			"stripe_checkout_session_id",
			"stripe_customer_id",
			"stripe_subscription_id",
		).
		Values(
			sub.ID,
			sub.AppID,
			sub.StripeCheckoutSessionID,
			sub.StripeCustomerID,
			sub.StripeSubscriptionID,
		),
	)
	if err != nil {
		return err
	}

	return nil
}
