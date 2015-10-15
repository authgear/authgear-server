package fs

import (
	"os"
	"testing"

	"github.com/oursky/skygear/oddb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDeviceQuery(t *testing.T) {
	Convey("Database", t, func() {
		dir := tempDir()
		defer os.RemoveAll(dir)

		db := newDeviceDatabase(dir)
		device0 := oddb.Device{
			ID:         "device0",
			Type:       "ios",
			Token:      "abcdef",
			UserInfoID: "user0",
		}
		db.Save(&device0)
		device1 := oddb.Device{
			ID:         "device1",
			Type:       "ios",
			Token:      "abcdef",
			UserInfoID: "user0",
		}
		db.Save(&device1)

		Convey("query user", func() {
			devices, err := db.Query("user0")
			So(err, ShouldBeNil)

			So(len(devices), ShouldEqual, 2)
			deviceids := []string{}
			for _, deviceinfo := range devices {
				deviceids = append(deviceids, deviceinfo.ID)
			}
			So(deviceids, ShouldContain, "device0")
			So(deviceids, ShouldContain, "device1")
		})

		Convey("query non-existent user", func() {
			devices, err := db.Query("user1")
			So(err, ShouldBeNil)
			So(len(devices), ShouldEqual, 0)
		})
	})
}
