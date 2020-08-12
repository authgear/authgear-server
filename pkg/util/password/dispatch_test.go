package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testHash struct {
	id             string
	testedPassword []byte
}

func (h *testHash) ID() string {
	return h.id
}
func (h *testHash) Hash(password []byte) ([]byte, error) {
	return constructPasswordFormat([]byte(h.id), password), nil
}
func (h *testHash) Compare(password, hash []byte) error {
	h.testedPassword = password
	return nil
}

func TestDispatch(t *testing.T) {
	Convey("Dispatching functions", t, func() {
		hash1 := testHash{id: "test1"}
		hash2 := testHash{id: "test2"}

		oldLatestFormat := latestFormat
		oldDefaultFormat := defaultFormat
		oldSupportedFormats := supportedFormats
		defer func() {
			latestFormat = oldLatestFormat
			defaultFormat = oldDefaultFormat
			supportedFormats = oldSupportedFormats
		}()
		latestFormat = &hash2
		defaultFormat = &hash1
		supportedFormats = map[string]passwordFormat{hash2.ID(): &hash2}

		Convey("should hash using correct format", func() {
			h, err := Hash([]byte("password"))
			So(err, ShouldBeNil)
			So(string(h), ShouldEqual, "$test2$password")
		})

		Convey("should compare using correct format", func() {
			var err error

			err = Compare([]byte("password1"), []byte("$test2$password1"))
			So(err, ShouldBeNil)
			So(string(hash2.testedPassword), ShouldEqual, "password1")

			err = Compare([]byte("password2"), []byte("$unk$password2"))
			So(err, ShouldBeNil)
			So(string(hash1.testedPassword), ShouldEqual, "password2")
		})

		Convey("should perform migration if needed", func() {
			h := []byte("$test1$password")
			migrated, err := TryMigrate([]byte("password"), &h)
			So(err, ShouldBeNil)
			So(migrated, ShouldBeTrue)
			So(string(h), ShouldEqual, "$test2$password")
		})

		Convey("should not perform migration if not needed", func() {
			h := []byte("$test2$password")
			migrated, err := TryMigrate([]byte("password"), &h)
			So(err, ShouldBeNil)
			So(migrated, ShouldBeFalse)
			So(string(h), ShouldEqual, "$test2$password")
		})
	})
}
