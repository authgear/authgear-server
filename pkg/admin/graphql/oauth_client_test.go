package graphql_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestParseDynamicClientsSourceArg(t *testing.T) {
	Convey("ParseDynamicClientsSourceArg", t, func() {
		Convey("an absent source means no filter", func() {
			source, err := graphql.ParseDynamicClientsSourceArg(map[string]any{})
			So(err, ShouldBeNil)
			So(source, ShouldBeNil)
		})

		Convey("an explicit null means no filter", func() {
			source, err := graphql.ParseDynamicClientsSourceArg(map[string]any{
				"source": nil,
			})
			So(err, ShouldBeNil)
			So(source, ShouldBeNil)
		})

		Convey("DCR filters by DCR", func() {
			source, err := graphql.ParseDynamicClientsSourceArg(map[string]any{
				"source": string(model.OAuthClientSourceDCR),
			})
			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)
			So(*source, ShouldEqual, model.OAuthClientSourceDCR)
		})

		Convey("CIMD filters by CIMD", func() {
			source, err := graphql.ParseDynamicClientsSourceArg(map[string]any{
				"source": string(model.OAuthClientSourceCIMD),
			})
			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)
			So(*source, ShouldEqual, model.OAuthClientSourceCIMD)
		})

		// An empty page would read as "this project has no static clients",
		// which is a different claim from "you cannot ask that here".
		Convey("STATIC is rejected rather than answered with an empty page", func() {
			source, err := graphql.ParseDynamicClientsSourceArg(map[string]any{
				"source": string(model.OAuthClientSourceStatic),
			})
			So(source, ShouldBeNil)
			So(err, ShouldBeError, graphql.ErrDynamicClientsSourceStatic)
		})
	})
}
