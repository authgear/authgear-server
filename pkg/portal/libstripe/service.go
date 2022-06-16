package libstripe

import (
	"context"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("stripe")} }

func NewClientAPI(stripeConfig *portalconfig.StripeConfig, logger Logger) *client.API {
	clientAPI := &client.API{}
	clientAPI.Init(stripeConfig.SecretKey, &stripe.Backends{
		API: stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
			LeveledLogger: logger,
		}),
	})
	return clientAPI
}

type PlanService interface {
	ListPlans() ([]*model.Plan, error)
}

type Service struct {
	ClientAPI *client.API
	Logger    Logger
	Context   context.Context
	Plans     PlanService
}

func (s *Service) FetchSubscriptionPlans() ([]*SubscriptionPlan, error) {
	plans, err := s.Plans.ListPlans()
	if err != nil {
		return nil, err
	}

	products, err := s.fetchProducts()
	if err != nil {
		return nil, err
	}
	subscriptionPlans, err := s.fetchPrices(plans, products)
	if err != nil {
		return nil, err
	}
	return subscriptionPlans, nil
}

func (s *Service) fetchProducts() ([]*stripe.Product, error) {
	var products []*stripe.Product

	listProductParams := &stripe.ProductListParams{
		ListParams: stripe.ListParams{
			Context: s.Context,
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

func (s *Service) fetchPrices(plans []*model.Plan, products []*stripe.Product) ([]*SubscriptionPlan, error) {
	knownPlanNames := make(map[string]struct{})
	for _, plan := range plans {
		knownPlanNames[plan.Name] = struct{}{}
	}

	m := make(map[string]*SubscriptionPlan)
	for _, product := range products {
		plan, ok := NewSubscriptionPlan(product, knownPlanNames)
		if ok {
			m[plan.StripeProductID] = plan
		}
	}

	listPriceParams := &stripe.PriceListParams{
		ListParams: stripe.ListParams{
			Context: s.Context,
		},
		Active: stripe.Bool(true),
	}
	iter := s.ClientAPI.Prices.List(listPriceParams)
	for iter.Next() {
		stripePrice := iter.Current().(*stripe.Price)
		productID := stripePrice.Product.ID
		if productPrice, ok := m[productID]; ok {
			price, err := NewPrice(stripePrice)
			if err != nil {
				return nil, err
			}
			productPrice.Prices = append(productPrice.Prices, price)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	var out []*SubscriptionPlan
	for _, productPrice := range m {
		out = append(out, productPrice)
	}

	return out, nil
}
