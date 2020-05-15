package interaction

import (
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

type mockOOBProvider struct {
	code int
}

func (p *mockOOBProvider) GenerateCode() string {
	code := p.code
	p.code++
	return strconv.Itoa(code)
}

func (p *mockOOBProvider) SendCode(opts oob.SendCodeOptions) error {
	return nil
}

func TestDoTriggerOOB(t *testing.T) {
	Convey("DoTriggerOOB", t, func() {
		timeProvider := &coretime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		p := &Provider{
			Time: timeProvider,
			OOB:  &mockOOBProvider{},
		}

		Convey("trigger first time", func() {
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

		Convey("trigger second time", func() {
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

			timeProvider.AdvanceSeconds(1)
			err = p.doTriggerOOB(i, action)
			So(err, ShouldEqual, ErrOOBOTPCooldown)

			timeProvider.AdvanceSeconds(59)
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
			timeProvider.AdvanceSeconds(1201)
			err = p.doTriggerOOB(i, action)
			So(err, ShouldBeNil)
			So(i.State[authenticator.AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPCode], ShouldEqual, "1")
			So(i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime], ShouldEqual, "2006-01-02T15:24:06Z")
			So(i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:24:06Z")
		})
	})
}
