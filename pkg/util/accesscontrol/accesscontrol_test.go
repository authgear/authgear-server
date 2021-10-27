package accesscontrol

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	Name   Subject = "/name"
	Gender Subject = "/gender"
)

const (
	EndUser   Role = "end_user"
	AdminUser Role = "admin_user"
)

const (
	Hidden Level = iota + 1
	Readonly
	Readwrite
)

func makeT() T {
	return T{
		Name: map[Role]Level{
			EndUser:   Hidden,
			AdminUser: Readwrite,
		},
		Gender: map[Role]Level{
			EndUser:   Readonly,
			AdminUser: Readwrite,
		},
	}
}

func TestGetLevel(t *testing.T) {
	Convey("GetLevel", t, func() {
		f := makeT().GetLevel

		So(f(Name, EndUser, 0), ShouldEqual, Hidden)
		So(f(Gender, EndUser, 0), ShouldEqual, Readonly)
		So(f(Name, AdminUser, 0), ShouldEqual, Readwrite)
		So(f(Gender, AdminUser, 0), ShouldEqual, Readwrite)
	})

	Convey("GetLevel default", t, func() {
		f := T{}.GetLevel

		So(f(Name, EndUser, 0), ShouldEqual, 0)
		So(f(Gender, EndUser, 0), ShouldEqual, 0)
		So(f(Name, AdminUser, 0), ShouldEqual, 0)
		So(f(Gender, AdminUser, 0), ShouldEqual, 0)
	})
}
