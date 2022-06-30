package libstripe

import (
	"fmt"
	"strconv"

	"github.com/stripe/stripe-go/v72"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

const (
	MetadataKeyAppID        = "app_id"
	MetadataKeyPlanName     = "plan_name"
	MetadataKeyPriceType    = "price_type"
	MetadataKeyUsageType    = "usage_type"
	MetadatakeySMSRegion    = "sms_region"
	MetadataKeyFreeQuantity = "free_quantity"
)

func NewPrice(stripeProduct *stripe.Product) (price *model.Price, err error) {
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

	price = &model.Price{
		StripeProductID: stripeProduct.ID,
		StripePriceID:   stripePrice.ID,
		Type:            priceType,
		Currency:        string(stripePrice.Currency),
		UnitAmount:      int(stripePrice.UnitAmount),
	}

	if stripePrice.TransformQuantity != nil {
		i := int(stripePrice.TransformQuantity.DivideBy)
		price.TransformQuantityDivideBy = &i
		price.TransformQuantityRound = model.TransformQuantityRound(stripePrice.TransformQuantity.Round)
	}

	if usageTypeStr, ok := stripeProduct.Metadata[MetadataKeyUsageType]; ok {
		usageType := model.UsageType(usageTypeStr)
		err = usageType.Valid()
		if err != nil {
			return
		}
		price.UsageType = usageType
	}

	if smsRegionStr, ok := stripeProduct.Metadata[MetadatakeySMSRegion]; ok {
		smsRegion := model.SMSRegion(smsRegionStr)
		err = smsRegion.Valid()
		if err != nil {
			return
		}
		price.SMSRegion = smsRegion
	}

	if freeQuantityStr, ok := stripePrice.Metadata[MetadataKeyFreeQuantity]; ok {
		var freeQuantity int
		freeQuantity, err = strconv.Atoi(freeQuantityStr)
		if err != nil {
			return
		}
		price.FreeQuantity = &freeQuantity
	}

	return
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
