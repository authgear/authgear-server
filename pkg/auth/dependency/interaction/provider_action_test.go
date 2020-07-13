package interaction

import (
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type mockOOBProvider struct {
	code int
}

func (p *mockOOBProvider) GenerateCode() string {
	code := p.code
	p.code++
	return strconv.Itoa(code)
}

func TestDoTriggerOOB(t *testing.T) {
	Convey("DoTriggerOOB", t, func() {
		clock := clock.NewMockClockAt("2006-01-02T15:04:05Z")
		p := &Provider{
			Clock: clock,
			OOB:   &mockOOBProvider{},
		}

		Convey("trigger first clock", func() {
			i := &Interaction{}
			spec := authenticator.Spec{
				Type: authn.AuthenticatorTypeOOB,
				Props: map[string]interface{}{
					authenticator.AuthenticatorPropOOBOTPID: "1",
				},
			}
			action := &ActionTriggerOOBAuthenticator{
				Authenticator: spec,
			}

			err := p.doTriggerOOB(i, action)
			So(err, ShouldBeNil)
			So(i.State[authenticator.AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime], ShouldEqual, "2006-01-02T15:04:05Z")
			So(i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:05Z")
		})

		Convey("trigger second clock", func() {
			i := &Interaction{}
			spec := authenticator.Spec{
				Type: authn.AuthenticatorTypeOOB,
				Props: map[string]interface{}{
					authenticator.AuthenticatorPropOOBOTPID: "1",
				},
			}
			action := &ActionTriggerOOBAuthenticator{
				Authenticator: spec,
			}

			err := p.doTriggerOOB(i, action)
			So(err, ShouldBeNil)
			So(i.State[authenticator.AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime], ShouldEqual, "2006-01-02T15:04:05Z")
			So(i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:05Z")

			clock.AdvanceSeconds(1)
			err = p.doTriggerOOB(i, action)
			So(err, ShouldEqual, ErrOOBOTPCooldown)

			clock.AdvanceSeconds(59)
			err = p.doTriggerOOB(i, action)
			So(err, ShouldBeNil)
			So(i.State[authenticator.AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime], ShouldEqual, "2006-01-02T15:04:05Z")
			So(i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:05:05Z")
		})

		Convey("generate new code", func() {
			i := &Interaction{}
			spec := authenticator.Spec{
				Type: authn.AuthenticatorTypeOOB,
				Props: map[string]interface{}{
					authenticator.AuthenticatorPropOOBOTPID: "1",
				},
			}
			action := &ActionTriggerOOBAuthenticator{
				Authenticator: spec,
			}

			err := p.doTriggerOOB(i, action)
			So(err, ShouldBeNil)
			So(i.State[authenticator.AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime], ShouldEqual, "2006-01-02T15:04:05Z")
			So(i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:05Z")

			// 20 minutes plus 1 second
			clock.AdvanceSeconds(1201)
			err = p.doTriggerOOB(i, action)
			So(err, ShouldBeNil)
			So(i.State[authenticator.AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPCode], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime], ShouldEqual, "2006-01-02T15:24:06Z")
			So(i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:24:06Z")
		})
	})
}
