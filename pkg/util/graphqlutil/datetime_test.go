package graphqlutil

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDateTimeInUTCFromInput(t *testing.T) {
	Convey("GetDateTimeInUTCFromInput", t, func() {
		Convey("should return nil when key does not exist", func() {
			input := map[string]interface{}{}
			result := GetDateTimeInUTCFromInput(input, "nonexistent")
			So(result, ShouldBeNil)
		})

		Convey("should return nil when value is nil", func() {
			input := map[string]interface{}{
				"timestamp": nil,
			}

			result := GetDateTimeInUTCFromInput(input, "timestamp")
			So(result, ShouldBeNil)
		})

		Convey("should handle time.Time value and convert to UTC", func() {
			// Create a time in a specific timezone (not UTC)
			originalTime := time.Date(2006, 1, 2, 3, 4, 5, 6, time.Local)

			input := map[string]interface{}{
				"timestamp": originalTime,
			}

			result := GetDateTimeInUTCFromInput(input, "timestamp")

			So(result, ShouldNotBeNil)
			So(result.Location(), ShouldEqual, time.UTC)
			// The time should be the same moment, just in UTC
			So(result.Unix(), ShouldEqual, originalTime.Unix())
		})

		Convey("should handle *time.Time value and convert to UTC", func() {
			// Create a time in a specific timezone (not UTC)
			originalTime := time.Date(2006, 1, 2, 3, 4, 5, 6, time.Local)

			input := map[string]interface{}{
				"timestamp": &originalTime,
			}

			result := GetDateTimeInUTCFromInput(input, "timestamp")

			So(result, ShouldNotBeNil)
			So(result.Location(), ShouldEqual, time.UTC)
			// The time should be the same moment, just in UTC
			So(result.Unix(), ShouldEqual, originalTime.Unix())
		})

		Convey("should handle time.Time already in UTC", func() {
			originalTime := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)

			input := map[string]interface{}{
				"timestamp": originalTime,
			}

			result := GetDateTimeInUTCFromInput(input, "timestamp")

			So(result, ShouldNotBeNil)
			So(result.Location(), ShouldEqual, time.UTC)
			So(result.Equal(originalTime), ShouldBeTrue)
		})

		Convey("should handle *time.Time already in UTC", func() {
			originalTime := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)

			input := map[string]interface{}{
				"timestamp": &originalTime,
			}

			result := GetDateTimeInUTCFromInput(input, "timestamp")

			So(result, ShouldNotBeNil)
			So(result.Location(), ShouldEqual, time.UTC)
			So(result.Equal(originalTime), ShouldBeTrue)
		})

		Convey("should panic with invalid type", func() {
			input := map[string]interface{}{
				"timestamp": "invalid-string",
			}

			So(func() {
				GetDateTimeInUTCFromInput(input, "timestamp")
			}, ShouldPanic)
		})

		Convey("should panic with integer value", func() {
			input := map[string]interface{}{
				"timestamp": 123456789,
			}

			So(func() {
				GetDateTimeInUTCFromInput(input, "timestamp")
			}, ShouldPanic)
		})
	})
}
