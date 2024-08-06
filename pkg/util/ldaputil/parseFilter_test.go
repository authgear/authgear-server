package ldaputil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseFilter(t *testing.T) {
	Convey("Parse Filter", t, func() {
		filterTemplate := "(&(objectclass=persion)(uid={{.Username}}))"
		username := "hi)(email=*"
		filter, _ := ParseFilter(filterTemplate, username)
		So(filter, ShouldNotEqual, "(&(objectclass=persion)(uid=hi)(email=*))")
	})
}
