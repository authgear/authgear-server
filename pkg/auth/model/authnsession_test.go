package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
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
		Convey("StepForward", func() {
			a := AuthnSession{}

			a.RequiredSteps = []AuthnSessionStep{"identity"}
			So(a.FinishedSteps, ShouldBeEmpty)
			a.StepForward()
			So(a.FinishedSteps, ShouldResemble, []AuthnSessionStep{"identity"})
			a.StepForward()
			So(a.FinishedSteps, ShouldResemble, []AuthnSessionStep{"identity"})

			a.RequiredSteps = []AuthnSessionStep{"identity", "mfa"}
			a.FinishedSteps = nil
			So(a.FinishedSteps, ShouldBeEmpty)
			a.StepForward()
			So(a.FinishedSteps, ShouldResemble, []AuthnSessionStep{"identity"})
			a.StepForward()
			So(a.FinishedSteps, ShouldResemble, []AuthnSessionStep{"identity", "mfa"})
			a.StepForward()
			So(a.FinishedSteps, ShouldResemble, []AuthnSessionStep{"identity", "mfa"})
		})
		Convey("Session", func() {
			a := AuthnSession{
				ClientID:                "client",
				UserID:                  "user",
				PrincipalID:             "principal",
				AuthenticatorID:         "authenticator",
				AuthenticatorType:       coreAuth.AuthenticatorTypeOOB,
				AuthenticatorOOBChannel: coreAuth.AuthenticatorOOBChannelSMS,
			}
			actual := a.Session()
			expected := coreAuth.Session{
				ClientID:                "client",
				UserID:                  "user",
				PrincipalID:             "principal",
				AuthenticatorID:         "authenticator",
				AuthenticatorType:       coreAuth.AuthenticatorTypeOOB,
				AuthenticatorOOBChannel: coreAuth.AuthenticatorOOBChannelSMS,
			}
			So(actual, ShouldResemble, expected)
		})
	})
}
