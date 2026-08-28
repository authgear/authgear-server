package oauthclient_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestStoreNewClient(t *testing.T) {
	Convey("Store.NewClient", t, func() {
		store := &oauthclient.Store{Clock: clock.NewMockClockAt("2026-08-17T00:00:00Z")}

		Convey("nil ClientName stays nil; DisplayName computes the fallback", func() {
			c := store.NewClient(&oauthclient.NewClientOptions{
				ClientID: "dcrc_test",
			})
			So(c.ClientName, ShouldBeNil)
			So(c.DisplayName(), ShouldEqual, "Client dcrc_test")
		})

		Convey(`empty-string ClientName is normalized to nil; DisplayName computes the fallback`, func() {
			empty := ""
			c := store.NewClient(&oauthclient.NewClientOptions{
				ClientID:   "dcrc_test",
				ClientName: &empty,
			})
			So(c.ClientName, ShouldBeNil)
			So(c.DisplayName(), ShouldEqual, "Client dcrc_test")
		})

		Convey("explicit ClientName is left untouched", func() {
			name := "PR #123 preview"
			c := store.NewClient(&oauthclient.NewClientOptions{
				ClientID:   "dcrc_test",
				ClientName: &name,
			})
			So(c.ClientName, ShouldNotBeNil)
			So(*c.ClientName, ShouldEqual, "PR #123 preview")
		})
	})
}
