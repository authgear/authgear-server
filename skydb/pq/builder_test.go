package pq

import (
	"testing"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPredicateSqlizerFactory(t *testing.T) {
	Convey("Functional Predicate", t, func() {
		Convey("user discover must be used with user record", func() {
			f := newPredicateSqlizerFactory(nil, "note")
			userDiscover := skydb.UserDiscoverFunc{
				Emails: []string{},
			}
			_, err := f.newUserDiscoverFunctionalPredicateSqlizer(userDiscover)
			So(err, ShouldNotBeNil)
		})
	})
}
