package libstripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

const (
	MetadataKeyAppID     = "app_id"
	MetadataKeyPlanName  = "plan_name"
	MetadataKeyPriceType = "price_type"
	MetadataKeyUsageType = "usage_type"
	MetadatakeySMSRegion = "sms_region"
)

type Price struct {
	StripePriceID   string          `json:"stripePriceID"`
	StripeProductID string          `json:"stripeProductID"`
	Currency        string          `json:"currency"`
	UnitAmount      int             `json:"unitAmount"`
	Type            model.PriceType `json:"type"`
	UsageType       model.UsageType `json:"usageType,omitempty"`
	SMSRegion       model.SMSRegion `json:"smsRegion,omitempty"`
}

func NewPrice(stripeProduct *stripe.Product) (price *Price, err error) {
	priceType := model.PriceType(stripeProduct.Metadata[MetadataKeyPriceType])
	err = priceType.Valid()
	if err != nil {
		return
	}

	stripePrice := stripeProduct.DefaultPrice
	if stripePrice == nil {
		err = fmt.Errorf("missing default price in the stripe product: %s", stripeProduct.Name)
		return
	}

	price = &Price{
		StripeProductID: stripeProduct.ID,
		StripePriceID:   stripePrice.ID,
		Type:            priceType,
		Currency:        string(stripePrice.Currency),
		UnitAmount:      int(stripePrice.UnitAmount),
	}

	switch priceType {
	case model.PriceTypeFixed:
		break
	case model.PriceTypeUsage:
		usageType := model.UsageType(stripeProduct.Metadata[MetadataKeyUsageType])
		err = usageType.Valid()
		if err != nil {
			return
		}
		smsRegion := model.SMSRegion(stripeProduct.Metadata[MetadatakeySMSRegion])
		err = smsRegion.Valid()
		if err != nil {
			return
		}
		price.UsageType = usageType
		price.SMSRegion = smsRegion
	}

	return
}

type SubscriptionPlan struct {
	Name   string   `json:"name"`
	Prices []*Price `json:"prices,omitempty"`
}

func NewSubscriptionPlan(planName string) *SubscriptionPlan {
	return &SubscriptionPlan{
		Name: planName,
	}
}

type CheckoutSession struct {
	StripeCheckoutSessionID string
	StripeCustomerID        *string
	AppID                   string
	URL                     string
	ExpiresAt               int64
	Status                  stripe.CheckoutSessionStatus
}

func (cs *CheckoutSession) IsCompleted() bool {
	return cs.Status == stripe.CheckoutSessionStatusComplete
}

func NewCheckoutSession(checkoutSession *stripe.CheckoutSession) *CheckoutSession {
	cs := &CheckoutSession{
		StripeCheckoutSessionID: checkoutSession.ID,
		AppID:                   checkoutSession.Metadata[MetadataKeyAppID],
		URL:                     checkoutSession.URL,
		ExpiresAt:               checkoutSession.ExpiresAt,
		Status:                  checkoutSession.Status,
	}

	if checkoutSession.Customer != nil {
		cs.StripeCustomerID = &checkoutSession.Customer.ID
	}

	return cs
}
