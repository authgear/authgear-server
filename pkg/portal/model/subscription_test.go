package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSubscriptionUsageItemMatch(t *testing.T) {
	Convey("SubscriptionUsageItem.Match", t, func() {
		So((&SubscriptionUsageItem{
			Type:      PriceTypeFixed,
			UsageType: UsageTypeNone,
			SMSRegion: SMSRegionNone,
		}).Match(&Price{
			Type:      PriceTypeFixed,
			UsageType: UsageTypeNone,
			SMSRegion: SMSRegionNone,
		}), ShouldBeTrue)

		So((&SubscriptionUsageItem{
			Type:      PriceTypeFixed,
			UsageType: UsageTypeNone,
			SMSRegion: SMSRegionNone,
		}).Match(&Price{
			Type:      PriceTypeUsage,
			UsageType: UsageTypeSMS,
			SMSRegion: SMSRegionNorthAmerica,
		}), ShouldBeFalse)

		So((&SubscriptionUsageItem{
			Type:      PriceTypeUsage,
			UsageType: UsageTypeSMS,
			SMSRegion: SMSRegionNorthAmerica,
		}).Match(&Price{
			Type:      PriceTypeUsage,
			UsageType: UsageTypeSMS,
			SMSRegion: SMSRegionNorthAmerica,
		}), ShouldBeTrue)

		So((&SubscriptionUsageItem{
			Type:      PriceTypeUsage,
			UsageType: UsageTypeSMS,
			SMSRegion: SMSRegionNorthAmerica,
		}).Match(&Price{
			Type:      PriceTypeUsage,
			UsageType: UsageTypeSMS,
			SMSRegion: SMSRegionOtherRegions,
		}), ShouldBeFalse)
	})
}

func TestSubscriptionUsageItemFillFrom(t *testing.T) {
	Convey("SubscriptionUsageItem.FillFrom", t, func() {
		Convey("quantity * unit amount", func() {
			So(*(&SubscriptionUsageItem{
				Quantity: 2,
			}).FillFrom(&Price{
				UnitAmount: 3,
			}).TotalAmount, ShouldEqual, 6)
		})

		Convey("max(0, quantity - free quantity) * unit amount", func() {
			freeQuantity := 10

			So(*(&SubscriptionUsageItem{
				Quantity: 2,
			}).FillFrom(&Price{
				FreeQuantity: &freeQuantity,
				UnitAmount:   3,
			}).TotalAmount, ShouldEqual, 0)

			So(*(&SubscriptionUsageItem{
				Quantity: 13,
			}).FillFrom(&Price{
				FreeQuantity: &freeQuantity,
				UnitAmount:   3,
			}).TotalAmount, ShouldEqual, 9)
		})

		Convey("ceil(quantity / divisor) * unit amount", func() {
			divisor := 5

			So(*(&SubscriptionUsageItem{
				Quantity: 2,
			}).FillFrom(&Price{
				TransformQuantityDivideBy: &divisor,
				TransformQuantityRound:    TransformQuantityRoundUp,
				UnitAmount:                3,
			}).TotalAmount, ShouldEqual, 3)

			So(*(&SubscriptionUsageItem{
				Quantity: 5,
			}).FillFrom(&Price{
				TransformQuantityDivideBy: &divisor,
				TransformQuantityRound:    TransformQuantityRoundUp,
				UnitAmount:                3,
			}).TotalAmount, ShouldEqual, 3)

			So(*(&SubscriptionUsageItem{
				Quantity: 6,
			}).FillFrom(&Price{
				TransformQuantityDivideBy: &divisor,
				TransformQuantityRound:    TransformQuantityRoundUp,
				UnitAmount:                3,
			}).TotalAmount, ShouldEqual, 6)
		})

		Convey("floor(quantity / divisor) * unit amount", func() {
			divisor := 5

			So(*(&SubscriptionUsageItem{
				Quantity: 2,
			}).FillFrom(&Price{
				TransformQuantityDivideBy: &divisor,
				TransformQuantityRound:    TransformQuantityRoundDown,
				UnitAmount:                3,
			}).TotalAmount, ShouldEqual, 0)

			So(*(&SubscriptionUsageItem{
				Quantity: 5,
			}).FillFrom(&Price{
				TransformQuantityDivideBy: &divisor,
				TransformQuantityRound:    TransformQuantityRoundDown,
				UnitAmount:                3,
			}).TotalAmount, ShouldEqual, 3)

			So(*(&SubscriptionUsageItem{
				Quantity: 6,
			}).FillFrom(&Price{
				TransformQuantityDivideBy: &divisor,
				TransformQuantityRound:    TransformQuantityRoundDown,
				UnitAmount:                3,
			}).TotalAmount, ShouldEqual, 3)
		})

		Convey("ceil(max(0, quantity - free quantity) / divisor) * unit amount", func() {
			freeQuantity := 10
			divisor := 5

			test := func(quantity int, expected int) {
				So(*(&SubscriptionUsageItem{
					Quantity: quantity,
				}).FillFrom(&Price{
					FreeQuantity:              &freeQuantity,
					TransformQuantityDivideBy: &divisor,
					TransformQuantityRound:    TransformQuantityRoundUp,
					UnitAmount:                3,
				}).TotalAmount, ShouldEqual, expected)
			}

			test(0, 0)
			test(1, 0)
			test(2, 0)
			test(3, 0)
			test(4, 0)
			test(5, 0)
			test(6, 0)
			test(7, 0)
			test(8, 0)
			test(9, 0)
			test(10, 0)
			test(11, 3)
			test(12, 3)
			test(13, 3)
			test(14, 3)
			test(15, 3)
			test(16, 6)
		})
	})
}
