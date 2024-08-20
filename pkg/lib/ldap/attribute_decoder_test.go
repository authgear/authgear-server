package ldap

import (
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAttributeDecoder(t *testing.T) {
	Convey("StringDecoder", t, func() {
		decoder := StringAttributeDecoder{}

		str := "Hello World"
		bytes := []byte(str)

		values, err := decoder.DecodeToStringRepresentable(bytes)
		So(err, ShouldBeNil)
		So(values, ShouldEqual, str)
	})

	Convey("UUIDDecoder", t, func() {
		decoder := UUIDAttributeDecoder{}

		UUID := uuid.Must(uuid.NewRandom())
		uuidBytes, err := UUID.MarshalBinary()
		So(err, ShouldBeNil)

		values, err := decoder.DecodeToStringRepresentable(uuidBytes)
		So(err, ShouldBeNil)
		So(values, ShouldEqual, UUID.String())
	})
}
