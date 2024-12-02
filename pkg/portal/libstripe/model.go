package libstripe

import (
	"fmt"
	"strconv"

	"github.com/stripe/stripe-go/v72"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

const (
	MetadataKeyAppID          = "app_id"
	MetadataKeyPlanName       = "plan_name"
	MetadataKeyVersion        = "version"
	MetadataKeyPriceType      = "price_type"
	MetadataKeyUsageType      = "usage_type"
	MetadatakeySMSRegion      = "sms_region"
	MetadatakeyWhatsappRegion = "whatsapp_region"
	MetadataKeyFreeQuantity   = "free_quantity"
)

func NewPriceFromProductOfProductList(stripeProduct *stripe.Product) (price *model.Price, err error) {
	stripePrice := stripeProduct.DefaultPrice
	if stripePrice == nil {
		err = fmt.Errorf("missing default price in the stripe product: %s", stripeProduct.Name)
		return
	}
	return newPrice(stripeProduct, stripePrice)
}

func NewPriceFromProductOfSubscription(stripeProduct *stripe.Product, stripePrice *stripe.Price) (price *model.Price, err error) {
	return newPrice(stripeProduct, stripePrice)
}

func newPrice(stripeProduct *stripe.Product, stripePrice *stripe.Price) (price *model.Price, err error) {
	priceType := model.PriceType(stripeProduct.Metadata[MetadataKeyPriceType])
	err = priceType.Valid()
	if err != nil {
		return
	}

	price = &model.Price{
		StripeProductID: stripeProduct.ID,
		StripePriceID:   stripePrice.ID,
		Type:            priceType,
		Currency:        string(stripePrice.Currency),
	}

	if len(stripePrice.Tiers) > 0 {
		firstTier := stripePrice.Tiers[0]
		if firstTier.UnitAmount == 0 {
			// If the first tier is free, set the amount to FreeQuanity
			freeQuantity := int(firstTier.UpTo)
			price.FreeQuantity = &freeQuantity
			if len(stripePrice.Tiers) > 1 {
				// And get the next tier as the unit price
				secondTier := stripePrice.Tiers[1]
				price.UnitAmount = int(secondTier.UnitAmount)
			}
		} else {
			price.UnitAmount = int(firstTier.UnitAmount)
		}
	} else {
		price.UnitAmount = int(stripePrice.UnitAmount)
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

	if whatsappRegionStr, ok := stripeProduct.Metadata[MetadatakeyWhatsappRegion]; ok {
		whatsappRegion := model.WhatsappRegion(whatsappRegionStr)
		err = whatsappRegion.Valid()
		if err != nil {
			return
		}
		price.WhatsappRegion = whatsappRegion
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

type planWithVersion struct {
	PlanName string
	Version  string
}
