package model_test

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type stubClient struct {
	isDynamicClient bool
	isThirdParty    bool
}

func (s stubClient) IsDynamicClient() bool { return s.isDynamicClient }
func (s stubClient) IsThirdParty() bool    { return s.isThirdParty }

func TestAccessPolicy(t *testing.T) {
	Convey("AccessPolicy.AllowsClient", t, func() {
		// Full 4x4: for each of the four (isDynamicClient, isThirdParty)
		// pairs, a policy with exactly one field true must return true only
		// for its own pair. A transposed arm passes any test that only
		// checks the diagonal, so all sixteen combinations are asserted.
		pairs := []struct {
			isDynamicClient bool
			isThirdParty    bool
		}{
			{false, false}, // static first-party
			{false, true},  // static third-party
			{true, false},  // dynamic first-party
			{true, true},   // dynamic third-party
		}

		policies := []struct {
			name   string
			policy model.AccessPolicy
			// ownPair is the index into pairs that this policy's single
			// true field corresponds to.
			ownPair int
		}{
			{"static first-party", model.AccessPolicy{AllowStaticFirstPartyClientAccess: true}, 0},
			{"static third-party", model.AccessPolicy{AllowStaticThirdPartyClientAccess: true}, 1},
			{"dynamic first-party", model.AccessPolicy{AllowDynamicFirstPartyClientAccess: true}, 2},
			{"dynamic third-party", model.AccessPolicy{AllowDynamicThirdPartyClientAccess: true}, 3},
		}

		for _, p := range policies {
			Convey(p.name+" policy allows only its own pair", func() {
				for i, pair := range pairs {
					got := p.policy.AllowsClient(stubClient{pair.isDynamicClient, pair.isThirdParty})
					if i == p.ownPair {
						So(got, ShouldBeTrue)
					} else {
						So(got, ShouldBeFalse)
					}
				}
			})
		}

		Convey("the zero value allows nothing", func() {
			var policy model.AccessPolicy
			for _, pair := range pairs {
				So(policy.AllowsClient(stubClient{pair.isDynamicClient, pair.isThirdParty}), ShouldBeFalse)
			}
		})
	})

	Convey("AccessPolicy JSON keys", t, func() {
		Convey("marshals exactly the four expected keys when all true", func() {
			policy := model.AccessPolicy{
				AllowStaticFirstPartyClientAccess:  true,
				AllowStaticThirdPartyClientAccess:  true,
				AllowDynamicFirstPartyClientAccess: true,
				AllowDynamicThirdPartyClientAccess: true,
			}
			b, err := json.Marshal(policy)
			So(err, ShouldBeNil)

			var m map[string]any
			err = json.Unmarshal(b, &m)
			So(err, ShouldBeNil)

			So(m, ShouldResemble, map[string]any{
				"allow_static_first_party_client_access":  true,
				"allow_static_third_party_client_access":  true,
				"allow_dynamic_first_party_client_access": true,
				"allow_dynamic_third_party_client_access": true,
			})
		})

		Convey("omits every key when all false", func() {
			b, err := json.Marshal(model.AccessPolicy{})
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, "{}")
		})
	})
}
