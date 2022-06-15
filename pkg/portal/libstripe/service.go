package libstripe

import (
	"context"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
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

type Service struct {
	ClientAPI *client.API
	Logger    Logger
	Context   context.Context
}

func (s *Service) FetchSubscriptionPlans() ([]*SubscriptionPlan, error) {
	// TODO(stripe): Fetch _portal_plan and return an intersection of plans.
	products, err := s.fetchProducts()
	if err != nil {
		return nil, err
	}
	subscriptionPlans, err := s.fetchPrices(products)
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

func (s *Service) fetchPrices(products []*stripe.Product) ([]*SubscriptionPlan, error) {
	m := make(map[string]*SubscriptionPlan)
	for _, product := range products {
		plan, ok := NewSubscriptionPlan(product)
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
			if price, ok := NewPrice(stripePrice); ok {
				productPrice.Prices = append(productPrice.Prices, price)
			}
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
