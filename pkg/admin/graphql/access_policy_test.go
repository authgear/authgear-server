package graphql

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDecodeAccessPolicyInput(t *testing.T) {
	Convey("decodeAccessPolicyInput", t, func() {
		Convey("accessPolicy absent returns nil", func() {
			patch := decodeAccessPolicyInput(map[string]any{})
			So(patch, ShouldBeNil)
		})

		Convey("accessPolicy explicitly null returns nil", func() {
			patch := decodeAccessPolicyInput(map[string]any{"accessPolicy": nil})
			So(patch, ShouldBeNil)
		})

		Convey("accessPolicy: {} returns a non-nil patch with every field nil", func() {
			patch := decodeAccessPolicyInput(map[string]any{"accessPolicy": map[string]any{}})
			So(patch, ShouldNotBeNil)
			So(patch.AllowStaticFirstPartyClientAccess, ShouldBeNil)
			So(patch.AllowStaticThirdPartyClientAccess, ShouldBeNil)
			So(patch.AllowDynamicFirstPartyClientAccess, ShouldBeNil)
			So(patch.AllowDynamicThirdPartyClientAccess, ShouldBeNil)
		})

		Convey("only the keys present in the input are set on the patch", func() {
			patch := decodeAccessPolicyInput(map[string]any{
				"accessPolicy": map[string]any{
					"allowStaticFirstPartyClientAccess": true,
				},
			})
			So(patch, ShouldNotBeNil)
			So(patch.AllowStaticFirstPartyClientAccess, ShouldResemble, new(true))
			So(patch.AllowStaticThirdPartyClientAccess, ShouldBeNil)
			So(patch.AllowDynamicFirstPartyClientAccess, ShouldBeNil)
			So(patch.AllowDynamicThirdPartyClientAccess, ShouldBeNil)
		})

		Convey("a key present as false sets a non-nil pointer to false", func() {
			// This is the case the JSONB merge (Store.UpdateResource) depends
			// on: a field explicitly cleared must marshal as "key": false,
			// not be indistinguishable from an unset field.
			patch := decodeAccessPolicyInput(map[string]any{
				"accessPolicy": map[string]any{
					"allowDynamicThirdPartyClientAccess": false,
				},
			})
			So(patch, ShouldNotBeNil)
			So(patch.AllowDynamicThirdPartyClientAccess, ShouldNotBeNil)
			So(*patch.AllowDynamicThirdPartyClientAccess, ShouldBeFalse)
		})

		Convey("all four keys present are all set", func() {
			patch := decodeAccessPolicyInput(map[string]any{
				"accessPolicy": map[string]any{
					"allowStaticFirstPartyClientAccess":  true,
					"allowStaticThirdPartyClientAccess":  false,
					"allowDynamicFirstPartyClientAccess": true,
					"allowDynamicThirdPartyClientAccess": false,
				},
			})
			So(patch, ShouldNotBeNil)
			So(patch.AllowStaticFirstPartyClientAccess, ShouldResemble, new(true))
			So(patch.AllowStaticThirdPartyClientAccess, ShouldResemble, new(false))
			So(patch.AllowDynamicFirstPartyClientAccess, ShouldResemble, new(true))
			So(patch.AllowDynamicThirdPartyClientAccess, ShouldResemble, new(false))
		})
	})
}
