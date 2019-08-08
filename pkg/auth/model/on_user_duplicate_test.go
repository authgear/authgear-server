package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidOnUserDuplicateForSSO(t *testing.T) {
	Convey("Test IsValidOnUserDuplicateForSSO", t, func() {
		So(IsValidOnUserDuplicateForSSO(""), ShouldBeFalse)
		So(IsValidOnUserDuplicateForSSO("nonsense"), ShouldBeFalse)

		So(IsValidOnUserDuplicateForSSO(OnUserDuplicateDefault), ShouldBeTrue)
		So(IsValidOnUserDuplicateForSSO(OnUserDuplicateAbort), ShouldBeTrue)
		So(IsValidOnUserDuplicateForSSO(OnUserDuplicateMerge), ShouldBeTrue)
		So(IsValidOnUserDuplicateForSSO(OnUserDuplicateCreate), ShouldBeTrue)
	})
}

func TestIsAllowedOnUserDuplicate(t *testing.T) {
	Convey("Test IsAllowedOnUserDuplicate", t, func() {
		f := IsAllowedOnUserDuplicate

		merge := false
		create := false

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeFalse)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeFalse)

		merge = true

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeFalse)

		create = true

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeTrue)

		merge = false

		So(f(merge, create, OnUserDuplicateAbort), ShouldBeTrue)
		So(f(merge, create, OnUserDuplicateMerge), ShouldBeFalse)
		So(f(merge, create, OnUserDuplicateCreate), ShouldBeTrue)
	})
}
