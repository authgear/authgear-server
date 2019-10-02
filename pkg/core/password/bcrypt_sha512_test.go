package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBCryptSHA512(t *testing.T) {
	Convey("BCrypt-SHA512", t, func() {
		bcrypt := bcryptSHA512Password{}
		Convey("should hash as expected", func() {
			h, err := bcrypt.Hash([]byte("password"))
			So(err, ShouldBeNil)
			So(string(h), ShouldStartWith, "$bcrypt-sha512$")
		})
		Convey("should compare as expected", func() {
			var h []byte

			h = []byte("$bcrypt-sha512$$2a$10$qQybX3kAJNT2YqYRVmbfjO5EhgWm6vV4cWmXo2ZATAuBmCyJM2fKu")
			So(bcrypt.Compare([]byte("password"), h), ShouldBeNil)
			So(bcrypt.Compare([]byte("Password"), h), ShouldBeError)

			h = []byte("$bcrypt-sha512$$2a$10$zDrD87PynMMSNjLGHu0y.uBZ8qf8j75xiBYmAK9QPqk7zqFvLigQi")
			So(bcrypt.Compare([]byte("password12345678password12345678password12345678password12345678password12345678"), h), ShouldBeNil)
			So(bcrypt.Compare([]byte("Password12345678password12345678password12345678password12345678password12345678"), h), ShouldBeError)
			// should NOT truncate passwords
			So(bcrypt.Compare([]byte("password12345678password12345678password12345678password12345678password12345670"), h), ShouldBeError)
		})
	})
}
