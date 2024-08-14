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
		byteValues := [][]byte{[]byte(str)}

		values, err := decoder.DecodeToStringRepresentable(byteValues)
		So(err, ShouldBeNil)
		So(values, ShouldResemble, []string{str})
	})

	Convey("UUIDDecoder", t, func() {
		decoder := UUIDAttributeDecoder{}

		UUID := uuid.Must(uuid.NewRandom())
		uuidBytes, err := UUID.MarshalBinary()
		So(err, ShouldBeNil)
		byteValues := [][]byte{uuidBytes}

		values, err := decoder.DecodeToStringRepresentable(byteValues)
		So(err, ShouldBeNil)
		So(values, ShouldResemble, []string{UUID.String()})
	})
}
