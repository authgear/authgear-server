package libstripe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
	"github.com/stripe/stripe-go/v72/webhook"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/redisutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

const RedisCacheKeySubscriptionPlans = "cache:portal:subscription-plans"

var ServiceLogger = slogutil.NewLogger("stripe")

func NewClientAPI(stripeConfig *portalconfig.StripeConfig) *client.API {
	clientAPI := &client.API{}
	clientAPI.Init(stripeConfig.SecretKey, &stripe.Backends{
		API: stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
			// The interface does not accept context.
			// Thus we can only ask it to suppress logging by using stripe.LevelNull.
			LeveledLogger: &stripe.LeveledLogger{
				Level: stripe.LevelNull,
			},
		}),
	})
	return clientAPI
}

type PlanService interface {
	ListPlans(ctx context.Context) ([]*plan.Plan, error)
}

type Cache interface {
	Get(context.Context, redisutil.SimpleCmdable, redisutil.Item) ([]byte, error)
}

type EndpointsProvider interface {
	BillingEndpointURL(relayGlobalAppID string) *url.URL
	BillingRedirectEndpointURL(relayGlobalAppID string) *url.URL
}

type Service struct {
	ClientAPI         *client.API
	Plans             PlanService
	GlobalRedisHandle *globalredis.Handle
	Cache             Cache
	Clock             clock.Clock
	StripeConfig      *portalconfig.StripeConfig
	Endpoints         EndpointsProvider
}

func (s *Service) FetchSubscriptionPlans(ctx context.Context) (subscriptionPlans []*model.SubscriptionPlan, err error) {
	item := redisutil.Item{
		Key:        RedisCacheKeySubscriptionPlans,
		Expiration: duration.PerHour,
		Do:         s.fetchSubscriptionPlans,
	}

	err = s.GlobalRedisHandle.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := s.Cache.Get(ctx, conn, item)
		if err != nil {
			return err
		}
		err = json.Unmarshal(bytes, &subscriptionPlans)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}

	return
}

func (s *Service) CreateCheckoutSession(ctx context.Context, appID string, customerEmail string, subscriptionPlan *model.SubscriptionPlan) (*CheckoutSession, error) {
	relayGlobalAppID := relay.ToGlobalID("App", appID)
	billingPageURL := s.Endpoints.BillingEndpointURL(relayGlobalAppID).String()
	billingRedirectPageURL := s.Endpoints.BillingRedirectEndpointURL(relayGlobalAppID).String()
	successURL := fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", billingRedirectPageURL)
	cancelURL := billingPageURL

	params := &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Context: ctx,
			Metadata: map[string]string{
				MetadataKeyAppID:    appID,
				MetadataKeyPlanName: subscriptionPlan.Name,
			},
		},
		SuccessURL:         &successURL,
		CancelURL:          &cancelURL,
		Mode:               stripe.String(string(stripe.CheckoutSessionModeSetup)),
		PaymentMethodTypes: []*string{stripe.String(string(stripe.PaymentMethodTypeCard))},
		CustomerCreation:   stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways)),
		// Collect billing address for tax.
		// See https://docs.stripe.com/tax/customer-locations#supported-formats
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionRequired)),
	}

	if customerEmail != "" {
		// If the customer email is empty
		// The customer will be asked to enter their email address during the checkout process
		params.CustomerEmail = &customerEmail
	}

	checkoutSession, err := s.ClientAPI.CheckoutSessions.New(params)
	if err != nil {
		return nil, err
	}

	return NewCheckoutSession(checkoutSession), nil
}

func (s *Service) GetSubscriptionPlan(ctx context.Context, planName string) (*model.SubscriptionPlan, error) {
	subscriptionPlans, err := s.FetchSubscriptionPlans(ctx)
	if err != nil {
		return nil, err
	}
	return s.getSubscriptionPlan(planName, subscriptionPlans)
}

func (s *Service) FetchCheckoutSession(ctx context.Context, checkoutSessionID string) (*CheckoutSession, error) {
	checkoutSession, err := s.ClientAPI.CheckoutSessions.Get(checkoutSessionID, &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Context: ctx,
		},
	})
	if err != nil {
		return nil, err
	}

	return NewCheckoutSession(checkoutSession), nil
}

func (s *Service) ConstructEvent(ctx context.Context, r *http.Request) (Event, error) {
	logger := ServiceLogger.GetLogger(ctx)
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	sig := r.Header.Get("Stripe-Signature")
	stripeEvent, err := webhook.ConstructEvent(payload, sig, s.StripeConfig.WebhookSigningKey)
	if err != nil {
		return nil, err
	}

	event, err := s.constructEvent(&stripeEvent)
	if errors.Is(err, ErrUnknownEvent) {
		logger.Info(ctx, "unhandled event", slog.String("payload", string(payload)))
	}
	return event, err
}

func (s *Service) CreateSubscriptionIfNotExists(ctx context.Context, checkoutSessionID string, subscriptionPlans []*model.SubscriptionPlan) error {
	// Fetch the checkout session
	expandSetupIntentPaymentMethod := "setup_intent.payment_method"
	expandCustomerSubscriptions := "customer.subscriptions"
	checkoutSession, err := s.ClientAPI.CheckoutSessions.Get(checkoutSessionID, &stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{&expandSetupIntentPaymentMethod, &expandCustomerSubscriptions},
		},
	})
	if err != nil {
		return err
	}

	planName := checkoutSession.Metadata[MetadataKeyPlanName]
	appID := checkoutSession.Metadata[MetadataKeyAppID]

	// Find the subscription plan
	subscriptionPlan, err := s.getSubscriptionPlan(planName, subscriptionPlans)
	if err != nil {
		return err
	}

	// Update invoice settings default
	customerID := &checkoutSession.Customer.ID
	pm := checkoutSession.SetupIntent.PaymentMethod
	customerParams := &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
		},
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.ID),
		},
	}

	_, err = s.ClientAPI.Customers.Update(*customerID, customerParams)
	if err != nil {
		return fmt.Errorf("failed to update customer default payment method: %w", err)
	}

	// Check if the custom has subscription to avoid duplicate subscription
	if checkoutSession.Customer.Subscriptions != nil && len(checkoutSession.Customer.Subscriptions.Data) > 0 {
		return ErrCustomerAlreadySubscribed
	}

	// Check if the app has subscription to avoid duplicate subscription
	// It was observed that in test mode, the following search query INCLUDES
	// subscriptions with status=canceled.
	// Therefore we have to check the actual result.
	hasSubscription := false
	iter := s.ClientAPI.Subscriptions.Search(&stripe.SubscriptionSearchParams{
		SearchParams: stripe.SearchParams{
			Context: ctx,
			Query:   fmt.Sprintf("status:'active' AND metadata['app_id']: '%s'", appID),
		},
	})
	for iter.Next() {
		sub := iter.Current().(*stripe.Subscription)
		if sub.Status == stripe.SubscriptionStatusActive && sub.Metadata[MetadataKeyAppID] == appID {
			hasSubscription = true
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to search app's subscription: %w", err)
	}
	if hasSubscription {
		return ErrAppAlreadySubscribed
	}

	// Create subscription
	subscriptionItems := []*stripe.SubscriptionItemsParams{}
	for _, p := range subscriptionPlan.Prices {
		subscriptionItems = append(subscriptionItems, &stripe.SubscriptionItemsParams{
			Price: stripe.String(p.StripePriceID),
		})
	}

	billingCycleAnchor := timeutil.FirstDayOfTheMonth(s.Clock.NowUTC()).AddDate(0, 1, 0)
	billingCycleAnchorUnix := billingCycleAnchor.Unix()
	_, err = s.ClientAPI.Subscriptions.New(&stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
			Metadata: map[string]string{
				MetadataKeyAppID:    appID,
				MetadataKeyPlanName: planName,
			},
		},
		Customer:           customerID,
		Items:              subscriptionItems,
		BillingCycleAnchor: &billingCycleAnchorUnix,
		AutomaticTax: &stripe.SubscriptionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SetSubscriptionCancelAtPeriodEnd(stripeSubscriptionID string, cancelAtPeriodEnd bool) (*time.Time, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(cancelAtPeriodEnd),
	}
	subscription, err := s.ClientAPI.Subscriptions.Update(stripeSubscriptionID, params)
	if err != nil {
		return nil, err
	}
	periodEnd := time.Unix(subscription.CurrentPeriodEnd, 0).UTC()
	return &periodEnd, nil
}

func (s *Service) fetchSubscriptionPlans(ctx context.Context) ([]byte, error) {
	plans, err := s.Plans.ListPlans(ctx)
	if err != nil {
		return nil, err
	}

	products, err := s.fetchProducts(ctx)
	if err != nil {
		return nil, err
	}

	knownPlansWithVersion := s.intersectPlanNames(plans, products)
	subscriptionPlans, err := s.convertToSubscriptionPlans(knownPlansWithVersion, products)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(subscriptionPlans)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (s *Service) fetchProducts(ctx context.Context) ([]*stripe.Product, error) {
	var products []*stripe.Product

	expandDefaultPrice := "data.default_price"
	expandTiers := "data.default_price.tiers"
	listProductParams := &stripe.ProductListParams{
		ListParams: stripe.ListParams{
			Context: ctx,
			Expand:  []*string{&expandDefaultPrice, &expandTiers},
		},
		Active: stripe.Bool(true),
	}
	iter := s.ClientAPI.Products.List(listProductParams)
	for iter.Next() {
		product := iter.Current().(*stripe.Product)
		products = append(products, product)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *Service) intersectPlanNames(plans []*plan.Plan, products []*stripe.Product) []planWithVersion {
	// plans can contain free plan that is not a paid plan.
	// products do not contain non-paid plan.
	// Therefore, we perform an intersection between the two.
	setA := make(map[string]struct{})
	for _, plan := range plans {
		setA[plan.Name] = struct{}{}
	}

	productPlanSet := make(map[string]planWithVersion)
	for _, product := range products {
		planName, planNameOk := product.Metadata[MetadataKeyPlanName]
		version := product.Metadata[MetadataKeyVersion]
		if planNameOk && planName != "" {
			planWithVersion := planWithVersion{
				PlanName: planName,
				Version:  version,
			}
			productPlanSet[planName] = planWithVersion
		}
	}

	intersection := setutil.Set[string]{}

	for a := range setA {
		_, ok := productPlanSet[a]
		if ok {
			intersection[a] = struct{}{}
		}
	}

	intersactedProductPlanWithVersions := []planWithVersion{}
	for _, planName := range intersection.Keys() {
		planWithVersion, ok := productPlanSet[planName]
		if !ok {
			panic(fmt.Errorf("unexpected: product of plan %v does not exist", planName))
		}
		intersactedProductPlanWithVersions = append(intersactedProductPlanWithVersions,
			planWithVersion,
		)
	}

	return intersactedProductPlanWithVersions
}

func (s *Service) convertToSubscriptionPlans(knownPlansWithVersion []planWithVersion, products []*stripe.Product) ([]*model.SubscriptionPlan, error) {
	m := make(map[string]*model.SubscriptionPlan)
	for _, it := range knownPlansWithVersion {
		m[it.PlanName] = model.NewSubscriptionPlan(it.PlanName, it.Version)
	}

	for _, product := range products {
		price, err := NewPriceFromProductOfProductList(product)
		if err != nil {
			// skip the unknown product
			continue
		}

		// If Product has plan name, then the product only applies to that plan.
		// Otherwise the product applies to every plan in the same version.
		planName := product.Metadata[MetadataKeyPlanName]
		version := product.Metadata[MetadataKeyVersion]
		if planName != "" {
			subscriptionPlan, ok := m[planName]
			// Tolerate product with unknown plan names.
			if ok {
				subscriptionPlan.Prices = append(subscriptionPlan.Prices, price)
			}
		} else {
			for _, subscriptionPlan := range m {
				if subscriptionPlan.Version != version {
					continue
				}
				subscriptionPlan.Prices = append(subscriptionPlan.Prices, price)
			}
		}
	}

	var out []*model.SubscriptionPlan
	for _, subscriptionPlan := range m {
		out = append(out, subscriptionPlan)
	}

	return out, nil
}

func (s *Service) getSubscriptionPlan(planName string, subscriptionPlans []*model.SubscriptionPlan) (*model.SubscriptionPlan, error) {
	var subscriptionPlan *model.SubscriptionPlan
	for _, sp := range subscriptionPlans {
		if sp.Name == planName {
			subscriptionPlan = sp
			break
		}
	}
	if subscriptionPlan == nil {
		return nil, fmt.Errorf("subscription plan not found")
	}

	return subscriptionPlan, nil
}

func (s *Service) constructEvent(stripeEvent *stripe.Event) (Event, error) {
	switch stripeEvent.Type {
	case string(EventTypeCheckoutSessionCompleted):
		object := stripeEvent.Data.Object
		checkoutSessionID, ok := object["id"].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		customerID, ok := object["customer"].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		metadata, ok := object["metadata"].(map[string]interface{})
		if !ok {
			return nil, ErrUnknownEvent
		}
		appID, ok := metadata[MetadataKeyAppID].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		planName, ok := metadata[MetadataKeyPlanName].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		return &CheckoutSessionCompletedEvent{
			AppID:                   appID,
			PlanName:                planName,
			StripeCheckoutSessionID: checkoutSessionID,
			StripeCustomerID:        customerID,
		}, nil
	case string(EventTypeCustomerSubscriptionCreated),
		string(EventTypeCustomerSubscriptionUpdated),
		string(EventTypeCustomerSubscriptionDeleted):
		object := stripeEvent.Data.Object
		subscriptionID, ok := object["id"].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}

		subscriptionStatus, ok := object["status"].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		customerID, ok := object["customer"].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		metadata, ok := object["metadata"].(map[string]interface{})
		if !ok {
			return nil, ErrUnknownEvent
		}
		appID, ok := metadata[MetadataKeyAppID].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		planName, ok := metadata[MetadataKeyPlanName].(string)
		if !ok {
			return nil, ErrUnknownEvent
		}
		if stripeEvent.Type == string(EventTypeCustomerSubscriptionCreated) {
			return &CustomerSubscriptionCreatedEvent{
				&CustomerSubscriptionEvent{
					StripeSubscriptionID:     subscriptionID,
					StripeCustomerID:         customerID,
					AppID:                    appID,
					PlanName:                 planName,
					StripeSubscriptionStatus: stripe.SubscriptionStatus(subscriptionStatus),
				},
			}, nil
		} else if stripeEvent.Type == string(EventTypeCustomerSubscriptionUpdated) {
			return &CustomerSubscriptionUpdatedEvent{
				&CustomerSubscriptionEvent{
					StripeSubscriptionID:     subscriptionID,
					StripeCustomerID:         customerID,
					AppID:                    appID,
					PlanName:                 planName,
					StripeSubscriptionStatus: stripe.SubscriptionStatus(subscriptionStatus),
				},
			}, nil
		} else if stripeEvent.Type == string(EventTypeCustomerSubscriptionDeleted) {
			return &CustomerSubscriptionDeletedEvent{
				&CustomerSubscriptionEvent{
					StripeSubscriptionID:     subscriptionID,
					StripeCustomerID:         customerID,
					AppID:                    appID,
					PlanName:                 planName,
					StripeSubscriptionStatus: stripe.SubscriptionStatus(subscriptionStatus),
				},
			}, nil
		}
		return nil, ErrUnknownEvent
	default:
		return nil, ErrUnknownEvent
	}
}

func (s *Service) GenerateCustomerPortalSession(appID string, customerID string) (*stripe.BillingPortalSession, error) {
	u := s.Endpoints.BillingEndpointURL(relay.ToGlobalID("App", appID))

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(u.String()),
	}

	return s.ClientAPI.BillingPortalSessions.New(params)
}

func (s *Service) UpdateSubscription(ctx context.Context, stripeSubscriptionID string, subscriptionPlan *model.SubscriptionPlan) (err error) {
	getParams := &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{stripe.String("items.data.price.product")},
		},
	}
	sub, err := s.ClientAPI.Subscriptions.Get(stripeSubscriptionID, getParams)
	if err != nil {
		return
	}

	// Update the plan name in metadata
	sub.Metadata[MetadataKeyPlanName] = subscriptionPlan.Name

	itemsParams, err := s.deriveSubscriptionItemsParams(sub, subscriptionPlan)
	if err != nil {
		return
	}

	updateParams := &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
			// Update metadata
			Metadata: sub.Metadata,
		},
		Items: itemsParams,
	}

	_, err = s.ClientAPI.Subscriptions.Update(stripeSubscriptionID, updateParams)
	if err != nil {
		return
	}

	return
}

func (s *Service) PreviewUpdateSubscription(ctx context.Context, stripeSubscriptionID string, subscriptionPlan *model.SubscriptionPlan) (preview *model.SubscriptionUpdatePreview, err error) {
	getParams := &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{stripe.String("items.data.price.product")},
		},
	}
	sub, err := s.ClientAPI.Subscriptions.Get(stripeSubscriptionID, getParams)
	if err != nil {
		return
	}

	itemsParams, err := s.deriveSubscriptionItemsParams(sub, subscriptionPlan)
	if err != nil {
		return
	}

	invoiceParams := &stripe.InvoiceParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Customer:          stripe.String(sub.Customer.ID),
		Subscription:      stripe.String(sub.ID),
		SubscriptionItems: itemsParams,
	}

	inv, err := s.ClientAPI.Invoices.GetNext(invoiceParams)
	if err != nil {
		return
	}

	preview = &model.SubscriptionUpdatePreview{
		Currency:  string(inv.Currency),
		AmountDue: int(inv.AmountDue),
	}
	return
}

func (s *Service) deriveSubscriptionItemsParams(sub *stripe.Subscription, subscriptionPlan *model.SubscriptionPlan) (out []*stripe.SubscriptionItemsParams, err error) {
	oldPrices, err := stripeSubscriptionToPrices(sub)
	if err != nil {
		return
	}
	newPrices := subscriptionPlan.Prices

	f := func(p *model.Price) string {
		return p.StripePriceID
	}
	oldPriceSet := setutil.NewSetFromSlice(oldPrices, f)
	newPriceSet := setutil.NewSetFromSlice(newPrices, f)

	pricesToBeRemoved := setutil.SetToSlice(
		oldPrices,
		oldPriceSet.Subtract(newPriceSet),
		f,
	)
	pricesToBeAdded := setutil.SetToSlice(
		newPrices,
		newPriceSet.Subtract(oldPriceSet),
		f,
	)

	for _, priceToBeRemoved := range pricesToBeRemoved {
		for _, item := range sub.Items.Data {
			if item.Price.ID == priceToBeRemoved.StripePriceID {
				out = append(out, &stripe.SubscriptionItemsParams{
					ID:         stripe.String(item.ID),
					Deleted:    stripe.Bool(true),
					ClearUsage: stripe.Bool(priceToBeRemoved.ShouldClearUsage()),
				})
			}
		}
	}
	for _, priceToBeAdded := range pricesToBeAdded {
		out = append(out, &stripe.SubscriptionItemsParams{
			Price: stripe.String(priceToBeAdded.StripePriceID),
		})
	}

	return
}

func (s *Service) GetSubscription(ctx context.Context, stripeCustomerID string) (*stripe.Subscription, error) {
	subscriptionListParams := &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Context: ctx,
			Expand: []*string{
				stripe.String("data.latest_invoice"),
				stripe.String("data.latest_invoice.payment_intent"),
			},
		},
		Customer: stripeCustomerID,
	}

	iter := s.ClientAPI.Subscriptions.List(subscriptionListParams)
	for iter.Next() {
		sub := iter.Current().(*stripe.Subscription)
		// Even the customer has more than 1 subscription,
		// we only consider the first one here.
		return sub, nil
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list subscription: %w", err)
	}
	return nil, ErrNoSubscription
}

func (s *Service) GetLastPaymentError(ctx context.Context, stripeCustomerID string) (*stripe.Error, error) {
	sub, err := s.GetSubscription(ctx, stripeCustomerID)
	if err != nil {
		if errors.Is(err, ErrNoSubscription) {
			// customer can have no subscription
			// e.g. right after the checkout session is created and
			// before the subscription is created
			return nil, nil
		}
		return nil, err
	}

	invoice := sub.LatestInvoice
	if invoice == nil {
		return nil, ErrNoInvoice
	}

	paymentIntent := invoice.PaymentIntent
	if paymentIntent == nil {
		return nil, ErrNoPaymentIntent
	}

	return paymentIntent.LastPaymentError, nil
}

// CancelSubscriptionImmediately removes the subscription immediately
// It should be used only for failed subscriptions
// To cancel normal subscription, SetSubscriptionCancelAtPeriodEnd should be used
func (s *Service) CancelSubscriptionImmediately(ctx context.Context, subscriptionID string) error {
	// By default, upon subscription cancellation, Stripe will stop automatic
	// collection of all finalized invoices for the customer. This is intended to
	// prevent unexpected payment attempts after the customer has canceled a subscription.
	//
	// https://stripe.com/docs/api/subscriptions/cancel
	params := &stripe.SubscriptionCancelParams{
		Params: stripe.Params{
			Context: ctx,
		},
	}
	_, err := s.ClientAPI.Subscriptions.Cancel(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}
	return nil
}

func stripeSubscriptionToPrices(subscription *stripe.Subscription) ([]*model.Price, error) {
	var prices []*model.Price
	for _, item := range subscription.Items.Data {
		stripePrice := item.Price
		stripeProduct := stripePrice.Product
		price, err := NewPriceFromProductOfSubscription(stripeProduct, stripePrice)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}
