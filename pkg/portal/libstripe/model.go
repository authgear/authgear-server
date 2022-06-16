package libstripe

import (
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

func (t PriceType) Valid() bool {
	switch t {
	case PriceTypeFixed:
		return true
	case PriceTypeUsage:
		return true
	}
	return false
}

type UsageType string

const (
	UsageTypeNone UsageType = ""
	UsageTypeSMS  UsageType = "sms"
)

func (t UsageType) Valid() bool {
	switch t {
	case UsageTypeNone:
		return true
	case UsageTypeSMS:
		return true
	}
	return false
}

type SMSRegion string

const (
	SMSRegionNone         SMSRegion = ""
	SMSRegionNorthAmerica SMSRegion = "north-america"
	SMSRegionOtherRegions SMSRegion = "other-regions"
)

func (r SMSRegion) Valid() bool {
	switch r {
	case SMSRegionNone:
		return true
	case SMSRegionNorthAmerica:
		return true
	case SMSRegionOtherRegions:
		return true
	}
	return false
}

type Price struct {
	StripePriceID string    `json:"stripePriceID"`
	Currency      string    `json:"currency"`
	UnitAmount    int       `json:"unitAmount"`
	Type          PriceType `json:"type"`
	UsageType     UsageType `json:"usageType,omitempty"`
	SMSRegion     SMSRegion `json:"smsRegion,omitempty"`
}

func NewPrice(stripePrice *stripe.Price) (price *Price, ok bool) {
	priceType := PriceType(stripePrice.Metadata[MetadataKeyPriceType])
	if !priceType.Valid() {
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
		if !usageType.Valid() {
			return
		}
		smsRegion := SMSRegion(stripePrice.Metadata[MetadatakeySMSRegion])
		if !smsRegion.Valid() {
			return
		}
		price.UsageType = usageType
		price.SMSRegion = smsRegion
	}

	ok = true
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
		return nil, false
	}
	return &SubscriptionPlan{
		StripeProductID: product.ID,
		Name:            planName,
	}, true
}
