package service

import (
	"database/sql"
	"errors"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/usage"
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
}

func (s *SubscriptionService) CreateSubscription(appID string, stripeSubscriptionID string, stripeCustomerID string) (*model.Subscription, error) {
	now := s.Clock.NowUTC()
	subscription := &model.Subscription{
		ID:                   uuid.New(),
		AppID:                appID,
		StripeCustomerID:     stripeCustomerID,
		StripeSubscriptionID: stripeSubscriptionID,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := s.createSubscription(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
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
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

// UpdateSubscriptionCheckoutStatus updates subscription checkout status and customer id
// It returns ErrSubscriptionCheckoutNotFound when the checkout is not found
// or the checkout status is already subscribed
func (s *SubscriptionService) UpdateSubscriptionCheckoutStatusAndCustomerID(appID string, stripCheckoutSessionID string, status model.SubscriptionCheckoutStatus, customerID string) error {
	now := s.Clock.NowUTC()
	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_portal_subscription_checkout")).
		Set("status", status).
		Set("stripe_customer_id", customerID).
		Set("updated_at", now).
		Where("stripe_checkout_session_id = ?", stripCheckoutSessionID).
		Where("app_id = ?", appID).
		// Only allow updating status if it is not subscribed
		Where("status != 'subscribed'")

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

func (s *SubscriptionService) UpdateSubscriptionCheckoutStatusByCustomerID(appID string, customerID string, status model.SubscriptionCheckoutStatus) error {
	now := s.Clock.NowUTC()
	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_portal_subscription_checkout")).
		Set("status", status).
		Set("updated_at", now).
		Where("app_id = ?", appID).
		Where("stripe_customer_id = ?", customerID).
		// Only allow updating status if it is not subscribed
		Where("status != 'subscribed'")

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

func (s *SubscriptionService) GetIsProcessingSubscription(appID string) (bool, error) {
	count, err := s.getCompletedSubscriptionCheckoutCount(appID)
	if err != nil {
		return false, err
	}

	hasSubscription := true
	_, err = s.GetSubscription(appID)
	if errors.Is(err, ErrSubscriptionNotFound) {
		hasSubscription = false
	} else if err != nil {
		return false, err
	}

	return count > 0 && !hasSubscription, nil
}

func (s *SubscriptionService) createSubscription(sub *model.Subscription) error {
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
		),
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

func (s *SubscriptionService) getCompletedSubscriptionCheckoutCount(appID string) (uint64, error) {
	query := s.SQLBuilder.
		Select("count(*)").
		From(s.SQLBuilder.TableName("_portal_subscription_checkout")).
		Where("app_id = ?", appID).
		Where("status = ?", model.SubscriptionCheckoutStatusCompleted)

	scan, err := s.SQLExecutor.QueryRowWith(query)
	if err != nil {
		return 0, err
	}

	var count uint64
	err = scan.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
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
		Type:      model.PriceTypeUsage,
		UsageType: model.UsageTypeSMS,
		SMSRegion: model.SMSRegionNorthAmerica,
		Quantity:  sumUsageRecord(rs1),
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
		Type:      model.PriceTypeUsage,
		UsageType: model.UsageTypeSMS,
		SMSRegion: model.SMSRegionOtherRegions,
		Quantity:  sumUsageRecord(rs2),
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
		Type:      model.PriceTypeUsage,
		UsageType: model.UsageTypeMAU,
		SMSRegion: model.SMSRegionNone,
		Quantity:  sumUsageRecord(rs3),
	}

	incompleteSubscriptionUsage := &model.SubscriptionUsage{
		NextBillingDate: stripeEnd,
		Items:           []*model.SubscriptionUsageItem{item1, item2, item3},
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
				Type:        price.Type,
				UsageType:   price.UsageType,
				SMSRegion:   price.SMSRegion,
				Quantity:    1,
				Currency:    &price.Currency,
				UnitAmount:  &price.UnitAmount,
				TotalAmount: &price.UnitAmount,
			})
		}
	}
}
