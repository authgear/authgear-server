package libstripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"
)

const (
	MetadataKeyAppID     = "app_id"
	MetadataKeyPlanName  = "plan_name"
	MetadataKeyPriceType = "price_type"
	MetadataKeyUsageType = "usage_type"
	MetadatakeySMSRegion = "sms_region"
)

type PriceType string

const (
	PriceTypeFixed PriceType = "fixed"
	PriceTypeUsage PriceType = "usage"
)

func (t PriceType) Valid() error {
	switch t {
	case PriceTypeFixed:
		return nil
	case PriceTypeUsage:
		return nil
	}
	return fmt.Errorf("stripe: unknown price_type: %#v", t)
}

type UsageType string

const (
	UsageTypeNone UsageType = ""
	UsageTypeSMS  UsageType = "sms"
)

func (t UsageType) Valid() error {
	switch t {
	case UsageTypeNone:
		return nil
	case UsageTypeSMS:
		return nil
	}
	return fmt.Errorf("stripe: unknown usage_type: %#v", t)
}

type SMSRegion string

const (
	SMSRegionNone         SMSRegion = ""
	SMSRegionNorthAmerica SMSRegion = "north-america"
	SMSRegionOtherRegions SMSRegion = "other-regions"
)

func (r SMSRegion) Valid() error {
	switch r {
	case SMSRegionNone:
		return nil
	case SMSRegionNorthAmerica:
		return nil
	case SMSRegionOtherRegions:
		return nil
	}
	return fmt.Errorf("stripe: unknown sms_region: %#v", r)
}

type Price struct {
	StripePriceID   string    `json:"stripePriceID"`
	StripeProductID string    `json:"stripeProductID"`
	Currency        string    `json:"currency"`
	UnitAmount      int       `json:"unitAmount"`
	Type            PriceType `json:"type"`
	UsageType       UsageType `json:"usageType,omitempty"`
	SMSRegion       SMSRegion `json:"smsRegion,omitempty"`
}

func NewPrice(stripeProduct *stripe.Product) (price *Price, err error) {
	priceType := PriceType(stripeProduct.Metadata[MetadataKeyPriceType])
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
	case PriceTypeFixed:
		break
	case PriceTypeUsage:
		usageType := UsageType(stripeProduct.Metadata[MetadataKeyUsageType])
		err = usageType.Valid()
		if err != nil {
			return
		}
		smsRegion := SMSRegion(stripeProduct.Metadata[MetadatakeySMSRegion])
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
