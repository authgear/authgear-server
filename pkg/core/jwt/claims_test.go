package jwt

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestClaimValidation(t *testing.T) {
	keyFunc := func(*jwt.Token) (interface{}, error) { return []byte("secret"), nil }
	Convey("StandardClaims", t, func() {
		Convey("Should validate without considering iat", func() {
			claims := StandardClaims{}
			_, err := jwt.ParseWithClaims("eyJhbGciOiJIUzI1NiJ9.e30.XmNK3GpH3Ys_7wsYBfq4C3M6goz71I7dTgUkuIa5lyQ", &claims, keyFunc)
			So(err, ShouldBeNil)
			So(claims.IssuedAt, ShouldEqual, 0)

			_, err = jwt.ParseWithClaims("eyJhbGciOiJIUzI1NiJ9.eyJpYXQiOjEwMH0.CEA8eZOVDQyZYdjLC4lRKCkhp7cMIFkfNfVlOOBNqns", &claims, keyFunc)
			So(err, ShouldBeNil)
			So(claims.IssuedAt, ShouldEqual, 100)
		})
	})
	Convey("MapClaims", t, func() {
		Convey("Should validate without considering iat", func() {
			claims := MapClaims{}
			_, err := jwt.ParseWithClaims("eyJhbGciOiJIUzI1NiJ9.e30.XmNK3GpH3Ys_7wsYBfq4C3M6goz71I7dTgUkuIa5lyQ", &claims, keyFunc)
			So(err, ShouldBeNil)
			So(claims["iat"], ShouldBeNil)

			_, err = jwt.ParseWithClaims("eyJhbGciOiJIUzI1NiJ9.eyJpYXQiOjEwMH0.CEA8eZOVDQyZYdjLC4lRKCkhp7cMIFkfNfVlOOBNqns", &claims, keyFunc)
			So(err, ShouldBeNil)
			So(claims["iat"], ShouldEqual, 100)
		})
	})
}
