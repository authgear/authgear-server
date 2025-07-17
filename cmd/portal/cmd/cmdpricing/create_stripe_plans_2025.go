package cmdpricing

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/cobrasentry"
)

var Dollar int64 = 100
var Cent int64 = 1

func createPlanProduct(ctx context.Context, api *client.API, planName string, planDisplayName string, price int64) error {
	logger := logger.GetLogger(ctx)

	params := &stripe.PriceParams{
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		UnitAmount:  stripe.Int64(price),
		TaxBehavior: stripe.String(string(stripe.PriceTaxBehaviorInclusive)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
		ProductData: &stripe.PriceProductDataParams{
			Name: stripe.String(planDisplayName),
			Metadata: map[string]string{
				"plan_name":              planName,
				"price_type":             "fixed",
				"subscription_item_type": "plan",
				"version":                "2025",
			},
		},
	}
	result, err := api.Prices.New(params)
	if err != nil {
		return err
	}

	updateProductParams := &stripe.ProductParams{
		DefaultPrice: stripe.String(result.ID),
	}
	_, err = api.Products.Update(result.Product.ID, updateProductParams)
	if err != nil {
		return err
	}

	logger.Info(ctx, "created product", slog.String("name", planDisplayName), slog.String("product_id", result.Product.ID))
	return nil
}

func createSMSProduct(ctx context.Context, api *client.API, displayName, itemType string, smsRegion string, perUnitPrice int64) error {
	logger := logger.GetLogger(ctx)

	params := &stripe.PriceParams{
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		TaxBehavior: stripe.String(string(stripe.PriceTaxBehaviorInclusive)),
		Recurring: &stripe.PriceRecurringParams{
			Interval:       stripe.String(string(stripe.PriceRecurringIntervalMonth)),
			UsageType:      stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
			AggregateUsage: stripe.String(string(stripe.PriceRecurringAggregateUsageSum)),
		},
		BillingScheme: stripe.String(string(stripe.PlanBillingSchemeTiered)),
		TiersMode:     stripe.String(string(stripe.PlanTiersModeGraduated)),
		Tiers: []*stripe.PriceTierParams{
			{
				UnitAmount: stripe.Int64(perUnitPrice),
				UpToInf:    stripe.Bool(true),
			},
		},
		ProductData: &stripe.PriceProductDataParams{
			Name: stripe.String(displayName),
			Metadata: map[string]string{
				"price_type":             "usage",
				"subscription_item_type": itemType,
				"sms_region":             smsRegion,
				"usage_type":             "sms",
				"version":                "2025",
			},
		},
	}
	result, err := api.Prices.New(params)
	if err != nil {
		return err
	}

	updateProductParams := &stripe.ProductParams{
		DefaultPrice: stripe.String(result.ID),
	}
	_, err = api.Products.Update(result.Product.ID, updateProductParams)
	if err != nil {
		return err
	}

	logger.Info(ctx, "created product", slog.String("name", displayName), slog.String("product_id", result.Product.ID))
	return nil
}

func createWhatsappProduct(ctx context.Context, api *client.API, displayName, itemType string, whatsappRegion string, perUnitPrice int64) error {
	logger := logger.GetLogger(ctx)

	params := &stripe.PriceParams{
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		TaxBehavior: stripe.String(string(stripe.PriceTaxBehaviorInclusive)),
		Recurring: &stripe.PriceRecurringParams{
			Interval:       stripe.String(string(stripe.PriceRecurringIntervalMonth)),
			UsageType:      stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
			AggregateUsage: stripe.String(string(stripe.PriceRecurringAggregateUsageSum)),
		},
		BillingScheme: stripe.String(string(stripe.PlanBillingSchemeTiered)),
		TiersMode:     stripe.String(string(stripe.PlanTiersModeGraduated)),
		Tiers: []*stripe.PriceTierParams{
			{
				UnitAmount: stripe.Int64(perUnitPrice),
				UpToInf:    stripe.Bool(true),
			},
		},
		ProductData: &stripe.PriceProductDataParams{
			Name: stripe.String(displayName),
			Metadata: map[string]string{
				"price_type":             "usage",
				"subscription_item_type": itemType,
				"whatsapp_region":        whatsappRegion,
				"usage_type":             "whatsapp",
				"version":                "2025",
			},
		},
	}
	result, err := api.Prices.New(params)
	if err != nil {
		return err
	}

	updateProductParams := &stripe.ProductParams{
		DefaultPrice: stripe.String(result.ID),
	}
	_, err = api.Products.Update(result.Product.ID, updateProductParams)
	if err != nil {
		return err
	}

	logger.Info(ctx, "created product", slog.String("name", displayName), slog.String("product_id", result.Product.ID))
	return nil
}

func createMAUProduct(ctx context.Context, api *client.API, displayName, planName string, perGroupPrice, groupUnit, freeQuantity int64) error {
	logger := logger.GetLogger(ctx)
	params := &stripe.PriceParams{
		Params: stripe.Params{
			Metadata: map[string]string{
				"free_quantity": fmt.Sprint(freeQuantity),
			},
		},
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		TaxBehavior: stripe.String(string(stripe.PriceTaxBehaviorInclusive)),
		Recurring: &stripe.PriceRecurringParams{
			Interval:       stripe.String(string(stripe.PriceRecurringIntervalMonth)),
			UsageType:      stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
			AggregateUsage: stripe.String(string(stripe.PriceRecurringAggregateUsageLastDuringPeriod)),
		},
		BillingScheme: stripe.String(string(stripe.PlanBillingSchemePerUnit)),
		UnitAmount:    stripe.Int64(perGroupPrice),
		TransformQuantity: &stripe.PriceTransformQuantityParams{
			DivideBy: &groupUnit,
			Round:    stripe.String(string(stripe.PriceTransformQuantityRoundUp)),
		},
		ProductData: &stripe.PriceProductDataParams{
			Name: stripe.String(displayName),
			Metadata: map[string]string{
				"plan_name":              planName,
				"price_type":             "usage",
				"subscription_item_type": "mau",
				"usage_type":             "mau",
				"version":                "2025",
			},
		},
	}
	result, err := api.Prices.New(params)
	if err != nil {
		return err
	}

	updateProductParams := &stripe.ProductParams{
		DefaultPrice: stripe.String(result.ID),
	}
	_, err = api.Products.Update(result.Product.ID, updateProductParams)
	if err != nil {
		return err
	}

	logger.Info(ctx, "created product", slog.String("name", displayName), slog.String("product_id", result.Product.ID))
	return nil
}

var cmdPricingCreateStripePlans2025 = &cobra.Command{
	Use: "create-stripe-plans-2025",
	RunE: cobrasentry.RunEWrap(portalcmd.GetBinder, func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		binder := portalcmd.GetBinder()

		stripeSecretKey, err := binder.GetRequiredString(cmd, portalcmd.ArgStripeSecretKey)
		if err != nil {
			return
		}

		stripeConfig := &portalconfig.StripeConfig{
			SecretKey: stripeSecretKey,
		}

		api := NewClientAPI(stripeConfig)

		err = createPlanProduct(ctx, api,
			"developers2025", "Developers (2025)",
			50*Dollar,
		)
		if err != nil {
			return err
		}

		err = createPlanProduct(ctx, api,
			"business2025", "Business (2025)",
			500*Dollar,
		)
		if err != nil {
			return err
		}

		err = createSMSProduct(ctx, api,
			"SMS usage (North America) (2025)",
			"sms-north-america",
			"north-america",
			2*Cent,
		)
		if err != nil {
			return err
		}

		err = createSMSProduct(ctx, api,
			"SMS usage (Other regions) (2025)",
			"sms-other-region",
			"other-regions",
			10*Cent,
		)
		if err != nil {
			return err
		}

		err = createWhatsappProduct(ctx, api,
			"Whatsapp Usage (North America) (2025)",
			"whatsapp-north-america",
			"north-america",
			2*Cent,
		)
		if err != nil {
			return err
		}

		err = createWhatsappProduct(ctx, api,
			"Whatsapp Usage (Other regions) (2025)",
			"whatsapp-other-region",
			"other-regions",
			10*Cent,
		)
		if err != nil {
			return err
		}

		err = createMAUProduct(ctx, api,
			"Business Plan Additional MAU (2025)",
			"business2025",
			50*Dollar, 5000,
			25000,
		)
		if err != nil {
			return err
		}

		return
	}),
}
