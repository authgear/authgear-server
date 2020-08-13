package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBCrypt(t *testing.T) {
	Convey("BCrypt", t, func() {
		bcrypt := bcryptPassword{}
		Convey("should hash as expected", func() {
			h, err := bcrypt.Hash([]byte("password"))
			So(err, ShouldBeNil)
			So(string(h), ShouldStartWith, "$2a$")
		})
		Convey("should compare as expected", func() {
			var h []byte

			h = []byte("$2a$10$4yzWhYLTp56Aire5CaS2EuUQjs0TiDa83faJe095mUeajNJUyrJDK")
			So(bcrypt.Compare([]byte("password"), h), ShouldBeNil)
			So(bcrypt.Compare([]byte("Password"), h), ShouldBeError)

			h = []byte("$2a$10$G8CkWW9JfhfLvlOugYKY7uepx4gPhdJm.QBLy2DqS3aJz7p7FOEjG")
			So(bcrypt.Compare([]byte("password12345678password12345678password12345678password12345678password12345678"), h), ShouldBeNil)
			So(bcrypt.Compare([]byte("Password12345678password12345678password12345678password12345678password12345678"), h), ShouldBeError)
			// bcrypt password truncation
			So(bcrypt.Compare([]byte("password12345678password12345678password12345678password12345678password12345670"), h), ShouldBeNil)
		})
	})
}
