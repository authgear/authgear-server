package service

import (
	"database/sql"
	"errors"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrSubscriptionCheckoutNotFound = apierrors.NotFound.WithReason("ErrSubscriptionCheckoutNotFound").
	New("subscription checkout not found")
var ErrSubscriptionNotFound = apierrors.NotFound.WithReason("ErrSubscriptionNotFound").New("subscription not found")

type SubscriptionConfigSourceStore interface {
	GetDatabaseSourceByAppID(appID string) (*configsource.DatabaseSource, error)
	UpdateDatabaseSource(dbs *configsource.DatabaseSource) error
}

type SubscriptionPlanStore interface {
	GetPlan(name string) (*model.Plan, error)
}

type UsageStore interface {
	FetchUploadedUsageRecords(
		appID string,
		recordName usage.RecordName,
		period periodical.Type,
		stripeStart time.Time,
		stripeEnd time.Time,
	) ([]*usage.UsageRecord, error)
	FetchUsageRecords(
		appID string,
		recordName usage.RecordName,
		period periodical.Type,
		startTime time.Time,
	) ([]*usage.UsageRecord, error)
}

type SubscriptionService struct {
	SQLBuilder        *globaldb.SQLBuilder
	SQLExecutor       *globaldb.SQLExecutor
	ConfigSourceStore SubscriptionConfigSourceStore
	PlanStore         SubscriptionPlanStore
	UsageStore        UsageStore
	Clock             clock.Clock
	AppConfig         *portalconfig.AppConfig
}

func (s *SubscriptionService) UpsertSubscription(appID string, stripeSubscriptionID string, stripeCustomerID string) (*model.Subscription, error) {
	now := s.Clock.NowUTC()
	if err := s.upsertSubscription(&model.Subscription{
		ID:                   uuid.New(),
		AppID:                appID,
		StripeCustomerID:     stripeCustomerID,
		StripeSubscriptionID: stripeSubscriptionID,
		CreatedAt:            now,
		UpdatedAt:            now,
	}); err != nil {
		return nil, err
	}
	return s.GetSubscription(appID)
}

func (s *SubscriptionService) CreateSubscriptionCheckout(checkoutSession *libstripe.CheckoutSession) (*model.SubscriptionCheckout, error) {
	now := s.Clock.NowUTC()
	cs := &model.SubscriptionCheckout{
		ID:                      uuid.New(),
		StripeCheckoutSessionID: checkoutSession.StripeCheckoutSessionID,
		AppID:                   checkoutSession.AppID,
		Status:                  model.SubscriptionCheckoutStatusOpen,
		CreatedAt:               now,
		UpdatedAt:               now,
		ExpireAt:                time.Unix(checkoutSession.ExpiresAt, 0).UTC(),
	}
	if err := s.createSubscriptionCheckout(cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (s *SubscriptionService) GetSubscription(appID string) (*model.Subscription, error) {
	q := s.SQLBuilder.Select(
		"id",
		"app_id",
		"stripe_customer_id",
		"stripe_subscription_id",
		"created_at",
		"updated_at",
		"cancelled_at",
		"ended_at",
	).
		From(s.SQLBuilder.TableName("_portal_subscription")).
		Where("app_id = ?", appID)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	var subscription model.Subscription
	err = row.Scan(
		&subscription.ID,
		&subscription.AppID,
		&subscription.StripeCustomerID,
		&subscription.StripeSubscriptionID,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
		&subscription.CancelledAt,
		&subscription.EndedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

// MarkCheckoutCompleted marks subscription checkout as completed.
// It returns ErrSubscriptionCheckoutNotFound when the checkout is not found
// or the checkout status is already subscribed.
func (s *SubscriptionService) MarkCheckoutCompleted(appID string, stripCheckoutSessionID string, customerID string) error {
	return s.updateSubscriptionCheckoutStatus(func(b squirrel.UpdateBuilder) squirrel.UpdateBuilder {
		return b.Set("status", model.SubscriptionCheckoutStatusCompleted).
			Set("stripe_customer_id", customerID).
			Where("stripe_checkout_session_id = ?", stripCheckoutSessionID).
			Where("app_id = ?", appID).
			// Only allow updating status if it is not subscribed
			Where("status != ?", model.SubscriptionCheckoutStatusSubscribed)
	})
}

// MarkCheckoutSubscribed marks subscription checkout as subscribed.
// It returns ErrSubscriptionCheckoutNotFound when the checkout is not found
// or the checkout status is already subscribed.
func (s *SubscriptionService) MarkCheckoutSubscribed(appID string, customerID string) error {
	return s.updateSubscriptionCheckoutStatus(func(b squirrel.UpdateBuilder) squirrel.UpdateBuilder {
		return b.Set("status", model.SubscriptionCheckoutStatusSubscribed).
			Where("app_id = ?", appID).
			Where("stripe_customer_id = ?", customerID).
			// Only allow updating status if it is not subscribed
			Where("status != ?", model.SubscriptionCheckoutStatusSubscribed)
	})
}

func (s *SubscriptionService) MarkCheckoutCancelled(appID string, customerID string) error {
	return s.updateSubscriptionCheckoutStatus(func(b squirrel.UpdateBuilder) squirrel.UpdateBuilder {
		return b.Set("status", model.SubscriptionCheckoutStatusCancelled).
			Where("app_id = ?", appID).
			Where("stripe_customer_id = ?", customerID)
	})
}

func (s *SubscriptionService) MarkCheckoutExpired(appID string, customerID string) error {
	return s.updateSubscriptionCheckoutStatus(func(b squirrel.UpdateBuilder) squirrel.UpdateBuilder {
		return b.Set("status", model.SubscriptionCheckoutStatusExpired).
			Where("app_id = ?", appID).
			Where("stripe_customer_id = ?", customerID)
	})
}

func (s *SubscriptionService) UpdateAppPlan(appID string, planName string) error {
	consrc, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(appID)
	if err != nil {
		return err
	}

	p, err := s.PlanStore.GetPlan(planName)
	if err != nil {
		return err
	}

	featureConfigYAML, err := yaml.Marshal(p.RawFeatureConfig)
	if err != nil {
		return err
	}

	consrc.PlanName = p.Name
	// json.Marshal handled base64 encoded of the YAML file
	consrc.Data[configsource.AuthgearFeatureYAML] = featureConfigYAML
	consrc.UpdatedAt = s.Clock.NowUTC()
	err = s.ConfigSourceStore.UpdateDatabaseSource(consrc)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) UpdateAppPlanToDefault(appID string) error {
	defaultPlan := s.AppConfig.DefaultPlan
	return s.UpdateAppPlan(appID, defaultPlan)
}

func (s *SubscriptionService) GetLastProcessingCustomerID(appID string) (*string, error) {
	hasSubscription := true
	_, err := s.GetSubscription(appID)
	if errors.Is(err, ErrSubscriptionNotFound) {
		hasSubscription = false
	} else if err != nil {
		return nil, err
	}

	// If the app has an active subscription, we ignore any processing checkout session.
	if hasSubscription {
		return nil, nil
	}

	return s.getLastCompletedSubscriptionCheckoutCustomerID(appID)
}

func (s *SubscriptionService) SetSubscriptionCancelledStatus(id string, cancelled bool, endedAt *time.Time) error {
	now := s.Clock.NowUTC()
	var subCancelledAt *time.Time
	var subEndedAt *time.Time
	if cancelled {
		subCancelledAt = &now
		subEndedAt = endedAt
	} else {
		subCancelledAt = nil
		subEndedAt = nil
	}

	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_portal_subscription")).
		Set("cancelled_at", subCancelledAt).
		Set("ended_at", subEndedAt).
		Set("updated_at", now).
		Where("id = ?", id)

	result, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSubscriptionCheckoutNotFound
	}

	return nil
}

func (s *SubscriptionService) ArchiveSubscription(sub *model.Subscription) error {
	now := s.Clock.NowUTC()
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_historical_subscription")).
		Columns(
			"id",
			"app_id",
			"stripe_customer_id",
			"stripe_subscription_id",
			"subscription_created_at",
			"subscription_updated_at",
			"subscription_cancelled_at",
			"subscription_ended_at",
			"created_at",
		).
		Values(
			uuid.New(),
			sub.AppID,
			sub.StripeCustomerID,
			sub.StripeSubscriptionID,
			sub.CreatedAt,
			sub.UpdatedAt,
			sub.CancelledAt,
			sub.EndedAt,
			now,
		),
	)
	if err != nil {
		return err
	}

	_, err = s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_subscription")).
		Where("id = ?", sub.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) upsertSubscription(sub *model.Subscription) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_subscription")).
		Columns(
			"id",
			"app_id",
			"stripe_customer_id",
			"stripe_subscription_id",
			"created_at",
			"updated_at",
		).
		Values(
			sub.ID,
			sub.AppID,
			sub.StripeCustomerID,
			sub.StripeSubscriptionID,
			sub.CreatedAt,
			sub.UpdatedAt,
		).Suffix("ON CONFLICT (app_id) DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id, stripe_subscription_id = EXCLUDED.stripe_subscription_id, updated_at = EXCLUDED.updated_at"),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) createSubscriptionCheckout(sc *model.SubscriptionCheckout) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_subscription_checkout")).
		Columns(
			"id",
			"app_id",
			"stripe_checkout_session_id",
			"stripe_customer_id",
			"status",
			"created_at",
			"updated_at",
			"expire_at",
		).
		Values(
			sc.ID,
			sc.AppID,
			sc.StripeCheckoutSessionID,
			sc.StripeCustomerID,
			sc.Status,
			sc.CreatedAt,
			sc.UpdatedAt,
			sc.ExpireAt,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) getLastCompletedSubscriptionCheckoutCustomerID(appID string) (*string, error) {
	now := s.Clock.NowUTC()
	query := s.SQLBuilder.
		Select("stripe_customer_id").
		From(s.SQLBuilder.TableName("_portal_subscription_checkout")).
		Where("app_id = ?", appID).
		Where("status = ?", model.SubscriptionCheckoutStatusCompleted).
		Where("expire_at > ?", now).
		OrderBy("created_at DESC").
		Limit(1)

	scan, err := s.SQLExecutor.QueryRowWith(query)
	if err != nil {
		return nil, err
	}

	var customerID string
	err = scan.Scan(&customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &customerID, nil
}

func (s *SubscriptionService) updateSubscriptionCheckoutStatus(
	f func(builder squirrel.UpdateBuilder) squirrel.UpdateBuilder,
) error {
	now := s.Clock.NowUTC()
	q := f(s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_portal_subscription_checkout")).
		Set("updated_at", now))

	result, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSubscriptionCheckoutNotFound
	}
	return nil
}

// GetSubscriptionUsage uses the current plan to estimate the usage and the cost.
// However, if we ever adjust the prices, the estimation will become inaccurate.
// A accurate estimation should use the Prices in the Stripe Subscription to perform calculation.
func (s *SubscriptionService) GetSubscriptionUsage(
	appID string,
	planName string,
	date time.Time,
	subscriptionPlans []*model.SubscriptionPlan,
) (*model.SubscriptionUsage, error) {
	firstDayOfMonth := timeutil.FirstDayOfTheMonth(date)
	stripeStart := firstDayOfMonth
	stripeEnd := stripeStart.AddDate(0, 1, 0)

	rs1, err := s.UsageStore.FetchUploadedUsageRecords(
		appID,
		usage.RecordNameSMSSentNorthAmerica,
		periodical.Daily,
		stripeStart,
		stripeEnd,
	)
	if err != nil {
		return nil, err
	}
	item1 := &model.SubscriptionUsageItem{
		Type:           model.PriceTypeUsage,
		UsageType:      model.UsageTypeSMS,
		SMSRegion:      model.SMSRegionNorthAmerica,
		WhatsappRegion: model.WhatsappRegionNone,
		Quantity:       sumUsageRecord(rs1),
	}

	rs2, err := s.UsageStore.FetchUploadedUsageRecords(
		appID,
		usage.RecordNameSMSSentOtherRegions,
		periodical.Daily,
		stripeStart,
		stripeEnd,
	)
	if err != nil {
		return nil, err
	}
	item2 := &model.SubscriptionUsageItem{
		Type:           model.PriceTypeUsage,
		UsageType:      model.UsageTypeSMS,
		SMSRegion:      model.SMSRegionOtherRegions,
		WhatsappRegion: model.WhatsappRegionNone,
		Quantity:       sumUsageRecord(rs2),
	}

	rs3, err := s.UsageStore.FetchUsageRecords(
		appID,
		usage.RecordNameActiveUser,
		periodical.Monthly,
		stripeStart,
	)
	if err != nil {
		return nil, err
	}
	item3 := &model.SubscriptionUsageItem{
		Type:           model.PriceTypeUsage,
		UsageType:      model.UsageTypeMAU,
		SMSRegion:      model.SMSRegionNone,
		WhatsappRegion: model.WhatsappRegionNone,
		Quantity:       sumUsageRecord(rs3),
	}

	rs4, err := s.UsageStore.FetchUploadedUsageRecords(
		appID,
		usage.RecordNameWhatsappSentNorthAmerica,
		periodical.Daily,
		stripeStart,
		stripeEnd,
	)
	if err != nil {
		return nil, err
	}
	item4 := &model.SubscriptionUsageItem{
		Type:           model.PriceTypeUsage,
		UsageType:      model.UsageTypeWhatsapp,
		SMSRegion:      model.SMSRegionNone,
		WhatsappRegion: model.WhatsappRegionNorthAmerica,
		Quantity:       sumUsageRecord(rs4),
	}

	rs5, err := s.UsageStore.FetchUploadedUsageRecords(
		appID,
		usage.RecordNameWhatsappSentOtherRegions,
		periodical.Daily,
		stripeStart,
		stripeEnd,
	)
	if err != nil {
		return nil, err
	}
	item5 := &model.SubscriptionUsageItem{
		Type:           model.PriceTypeUsage,
		UsageType:      model.UsageTypeWhatsapp,
		SMSRegion:      model.SMSRegionNone,
		WhatsappRegion: model.WhatsappRegionOtherRegions,
		Quantity:       sumUsageRecord(rs5),
	}

	incompleteSubscriptionUsage := &model.SubscriptionUsage{
		NextBillingDate: stripeEnd,
		Items:           []*model.SubscriptionUsageItem{item1, item2, item3, item4, item5},
	}

	targetPlan, ok := findPlan(planName, subscriptionPlans)
	if !ok {
		return incompleteSubscriptionUsage, nil
	}

	fillCost(incompleteSubscriptionUsage, targetPlan)
	return incompleteSubscriptionUsage, nil
}

func sumUsageRecord(records []*usage.UsageRecord) int {
	sum := 0
	for _, record := range records {
		sum += record.Count
	}
	return sum
}

func findPlan(planName string, subscriptionPlans []*model.SubscriptionPlan) (*model.SubscriptionPlan, bool) {
	// The first step is to find the plan.
	var targetPlan *model.SubscriptionPlan
	for _, plan := range subscriptionPlans {
		if plan.Name == planName {
			p := plan
			targetPlan = p
		}
	}
	if targetPlan == nil {
		return nil, false
	}
	return targetPlan, true
}

func fillCost(subscriptionUsage *model.SubscriptionUsage, subscriptionPlan *model.SubscriptionPlan) {
	// First fill in the cost of metered usage items.
	for _, item := range subscriptionUsage.Items {
		for _, price := range subscriptionPlan.Prices {
			if item.Match(price) {
				item.FillFrom(price)
			}
		}
	}

	// First of all, add an usage item to represent the fixed cost.
	for _, price := range subscriptionPlan.Prices {
		if price.Type == model.PriceTypeFixed {
			subscriptionUsage.Items = append(subscriptionUsage.Items, &model.SubscriptionUsageItem{
				Type:           price.Type,
				UsageType:      price.UsageType,
				SMSRegion:      price.SMSRegion,
				WhatsappRegion: price.WhatsappRegion,
				Quantity:       1,
				Currency:       &price.Currency,
				UnitAmount:     &price.UnitAmount,
				TotalAmount:    &price.UnitAmount,
			})
		}
	}
}
