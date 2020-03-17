package authn

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSession(t *testing.T) {
	Convey("authn.Session", t, func() {
		Convey("IsFinished", func() {
			a := Session{}
			So(a.IsFinished(), ShouldBeTrue)

			a.RequiredSteps = []SessionStep{"identity"}
			a.FinishedSteps = []SessionStep{}
			So(a.IsFinished(), ShouldBeFalse)

			a.RequiredSteps = []SessionStep{"identity"}
			a.FinishedSteps = []SessionStep{"identity"}
			So(a.IsFinished(), ShouldBeTrue)

			a.RequiredSteps = []SessionStep{"identity", "mfa"}
			a.FinishedSteps = []SessionStep{"identity"}
			So(a.IsFinished(), ShouldBeFalse)

			a.RequiredSteps = []SessionStep{"identity", "mfa"}
			a.FinishedSteps = []SessionStep{"identity", "mfa"}
			So(a.IsFinished(), ShouldBeTrue)
		})
		Convey("NextStep", func() {
			var step SessionStep
			var ok bool
			a := Session{}
			_, ok = a.NextStep()
			So(ok, ShouldBeFalse)

			a.RequiredSteps = []SessionStep{"identity"}
			a.FinishedSteps = []SessionStep{}
			step, ok = a.NextStep()
			So(ok, ShouldBeTrue)
			So(step, ShouldEqual, "identity")

			a.RequiredSteps = []SessionStep{"identity"}
			a.FinishedSteps = []SessionStep{"identity"}
			_, ok = a.NextStep()
			So(ok, ShouldBeFalse)

			a.RequiredSteps = []SessionStep{"identity", "mfa"}
			a.FinishedSteps = []SessionStep{"identity"}
			step, ok = a.NextStep()
			So(ok, ShouldBeTrue)
			So(step, ShouldEqual, "mfa")

			a.RequiredSteps = []SessionStep{"identity", "mfa"}
			a.FinishedSteps = []SessionStep{"identity", "mfa"}
			_, ok = a.NextStep()
			So(ok, ShouldBeFalse)
		})
	})
}
