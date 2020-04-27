package interaction

import (
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

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

func (p *mockOOBProvider) SendCode(spec AuthenticatorSpec, code string) error {
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
			spec := AuthenticatorSpec{
				ID:   "1",
				Type: AuthenticatorTypeOOBOTP,
			}
			action := &ActionTriggerOOBAuthenticator{
				Authenticator: spec,
			}

			p.doTriggerOOB(i, action)
			So(i.State[AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:05Z")
		})

		Convey("trigger second time", func() {
			i := &Interaction{}
			spec := AuthenticatorSpec{
				ID:   "1",
				Type: AuthenticatorTypeOOBOTP,
			}
			action := &ActionTriggerOOBAuthenticator{
				Authenticator: spec,
			}

			p.doTriggerOOB(i, action)
			So(i.State[AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:05Z")

			timeProvider.AdvanceSeconds(1)
			p.doTriggerOOB(i, action)
			So(i.State[AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:06Z")
		})

		Convey("switch authenticator", func() {
			i := &Interaction{}
			spec1 := AuthenticatorSpec{
				ID:   "1",
				Type: AuthenticatorTypeOOBOTP,
			}
			action := &ActionTriggerOOBAuthenticator{
				Authenticator: spec1,
			}

			p.doTriggerOOB(i, action)
			So(i.State[AuthenticatorStateOOBOTPID], ShouldEqual, "1")
			So(i.State[AuthenticatorStateOOBOTPCode], ShouldEqual, "0")
			So(i.State[AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:05Z")

			timeProvider.AdvanceSeconds(1)
			spec2 := AuthenticatorSpec{
				ID:   "2",
				Type: AuthenticatorTypeOOBOTP,
			}
			action = &ActionTriggerOOBAuthenticator{
				Authenticator: spec2,
			}

			p.doTriggerOOB(i, action)
			So(i.State[AuthenticatorStateOOBOTPID], ShouldEqual, "2")
			So(i.State[AuthenticatorStateOOBOTPCode], ShouldEqual, "1")
			So(i.State[AuthenticatorStateOOBOTPTriggerTime], ShouldEqual, "2006-01-02T15:04:06Z")
		})
	})
}
