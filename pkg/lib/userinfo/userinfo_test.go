package userinfo

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestUserInfoSerialization(t *testing.T) {
	Convey("UserInfo serialization", t, func() {
		u := &UserInfo{
			User: &model.User{
				StandardAttributes: map[string]interface{}{},
				CustomAttributes:   map[string]interface{}{},
				Web3:               map[string]interface{}{},
				Roles:              []string{},
				Groups:             []string{},
			},
			EffectiveRoleKeys: []string{},
		}

		b, err := json.Marshal(u)
		So(err, ShouldBeNil)

		var uu UserInfo
		err = json.Unmarshal(b, &uu)
		So(err, ShouldBeNil)

		So(uu.EffectiveRoleKeys, ShouldNotBeNil)
		So(uu.EffectiveRoleKeys, ShouldHaveLength, 0)

		So(uu.User.Roles, ShouldNotBeNil)
		So(uu.User.Roles, ShouldHaveLength, 0)

		So(uu.User.Groups, ShouldNotBeNil)
		So(uu.User.Groups, ShouldHaveLength, 0)

		So(uu.User.StandardAttributes, ShouldNotBeNil)
		So(uu.User.StandardAttributes, ShouldHaveLength, 0)

		So(uu.User.CustomAttributes, ShouldNotBeNil)
		So(uu.User.CustomAttributes, ShouldHaveLength, 0)
	})
}
