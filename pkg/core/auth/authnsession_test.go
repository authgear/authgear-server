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
	})
}
