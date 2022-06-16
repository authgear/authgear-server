package libstripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v72"
)

const (
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
	StripePriceID string    `json:"stripePriceID"`
	Currency      string    `json:"currency"`
	UnitAmount    int       `json:"unitAmount"`
	Type          PriceType `json:"type"`
	UsageType     UsageType `json:"usageType,omitempty"`
	SMSRegion     SMSRegion `json:"smsRegion,omitempty"`
}

func NewPrice(stripePrice *stripe.Price) (price *Price, err error) {
	// Though we tolerate unknown Products,
	// for known product, we do NOT tolerate unknown price.
	// If we were to tolerate unknown price,
	// we have a risk to present inaccurate pricing information.

	priceType := PriceType(stripePrice.Metadata[MetadataKeyPriceType])
	err = priceType.Valid()
	if err != nil {
		return
	}

	price = &Price{
		StripePriceID: stripePrice.ID,
		Type:          priceType,
		Currency:      string(stripePrice.Currency),
		UnitAmount:    int(stripePrice.UnitAmount),
	}

	switch priceType {
	case PriceTypeFixed:
		break
	case PriceTypeUsage:
		usageType := UsageType(stripePrice.Metadata[MetadataKeyUsageType])
		err = usageType.Valid()
		if err != nil {
			return
		}
		smsRegion := SMSRegion(stripePrice.Metadata[MetadatakeySMSRegion])
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
	StripeProductID string   `json:"stripeProductID"`
	Name            string   `json:"name"`
	Prices          []*Price `json:"prices,omitempty"`
}

func NewSubscriptionPlan(product *stripe.Product, knownPlanNames map[string]struct{}) (*SubscriptionPlan, bool) {
	planName := product.Metadata[MetadataKeyPlanName]
	_, ok := knownPlanNames[planName]
	if !ok {
		// There could exist some unknown Products on Stripe.
		// We tolerate that.
		return nil, false
	}
	return &SubscriptionPlan{
		StripeProductID: product.ID,
		Name:            planName,
	}, true
}
