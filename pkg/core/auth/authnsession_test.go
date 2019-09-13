package auth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthnSession(t *testing.T) {
	Convey("AuthnSession", t, func() {
		Convey("IsFinished", func() {
			a := AuthnSession{}
			So(a.IsFinished(), ShouldBeTrue)

			a.RequiredSteps = []AuthnSessionStep{"identity"}
			a.FinishedSteps = []AuthnSessionStep{}
			So(a.IsFinished(), ShouldBeFalse)

			a.RequiredSteps = []AuthnSessionStep{"identity"}
			a.FinishedSteps = []AuthnSessionStep{"identity"}
			So(a.IsFinished(), ShouldBeTrue)

			a.RequiredSteps = []AuthnSessionStep{"identity", "mfa"}
			a.FinishedSteps = []AuthnSessionStep{"identity"}
			So(a.IsFinished(), ShouldBeFalse)

			a.RequiredSteps = []AuthnSessionStep{"identity", "mfa"}
			a.FinishedSteps = []AuthnSessionStep{"identity", "mfa"}
			So(a.IsFinished(), ShouldBeTrue)
		})
		Convey("NextStep", func() {
			var step AuthnSessionStep
			var ok bool
			a := AuthnSession{}
			_, ok = a.NextStep()
			So(ok, ShouldBeFalse)

			a.RequiredSteps = []AuthnSessionStep{"identity"}
			a.FinishedSteps = []AuthnSessionStep{}
			step, ok = a.NextStep()
			So(ok, ShouldBeTrue)
			So(step, ShouldEqual, AuthnSessionStepIdentity)

			a.RequiredSteps = []AuthnSessionStep{"identity"}
			a.FinishedSteps = []AuthnSessionStep{"identity"}
			_, ok = a.NextStep()
			So(ok, ShouldBeFalse)

			a.RequiredSteps = []AuthnSessionStep{"identity", "mfa"}
			a.FinishedSteps = []AuthnSessionStep{"identity"}
			step, ok = a.NextStep()
			So(ok, ShouldBeTrue)
			So(step, ShouldEqual, AuthnSessionStepMFA)

			a.RequiredSteps = []AuthnSessionStep{"identity", "mfa"}
			a.FinishedSteps = []AuthnSessionStep{"identity", "mfa"}
			_, ok = a.NextStep()
			So(ok, ShouldBeFalse)
		})
		Convey("StepMFA", func() {
			a := AuthnSession{}
			step := func() error {
				return a.StepMFA(AuthnSessionStepMFAOptions{
					AuthenticatorID:   "authenticator",
					AuthenticatorType: AuthenticatorTypeTOTP,
				})
			}

			So(step(), ShouldBeError, "expected step to be mfa")

			a.RequiredSteps = []AuthnSessionStep{"identity"}
			So(step(), ShouldBeError, "expected step to be mfa")

			a.RequiredSteps = []AuthnSessionStep{"identity", "mfa"}
			a.FinishedSteps = []AuthnSessionStep{"identity"}
			So(step(), ShouldBeNil)
		})
		Convey("Session", func() {
			a := AuthnSession{
				ClientID:                "client",
				UserID:                  "user",
				PrincipalID:             "principal",
				PrincipalType:           "password",
				AuthenticatorID:         "authenticator",
				AuthenticatorType:       AuthenticatorTypeOOB,
				AuthenticatorOOBChannel: AuthenticatorOOBChannelSMS,
			}
			actual := a.Session()
			expected := Session{
				ClientID:                "client",
				UserID:                  "user",
				PrincipalID:             "principal",
				PrincipalType:           "password",
				AuthenticatorID:         "authenticator",
				AuthenticatorType:       AuthenticatorTypeOOB,
				AuthenticatorOOBChannel: AuthenticatorOOBChannelSMS,
			}
			So(actual, ShouldResemble, expected)
		})
	})
}
