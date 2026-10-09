package libstripe

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

func TestServiceFetchSubscriptionPlans(t *testing.T) {
	Convey("FetchSubscriptionPlans", t, func() {
		Convey("returns no plans without calling Stripe when Stripe is not configured", func() {
			s := &Service{
				StripeConfig: &portalconfig.StripeConfig{},
			}

			plans, err := s.FetchSubscriptionPlans(context.Background())
			So(err, ShouldBeNil)
			So(plans, ShouldNotBeNil)
			So(plans, ShouldBeEmpty)
		})
	})
}
