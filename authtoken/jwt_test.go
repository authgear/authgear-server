package authtoken

import (
	"errors"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestJWTStore(t *testing.T) {
	Convey("JWTStore", t, func() {
		store := NewJWTStore("secret", 0)

		Convey("should panic without secret", func() {
			So(func() { NewJWTStore("", 0) }, ShouldPanic)
		})

		Convey("should create new token", func() {
			token, err := store.NewToken("exampleapp", "userid1")
			So(err, ShouldBeNil)
			So(token.UserInfoID, ShouldEqual, "userid1")

			tokenString := token.AccessToken
			claims := jwt.StandardClaims{}
			jwtToken, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("incorrect signing method")
				}
				return []byte("secret"), nil
			})

			So(jwtToken.Valid, ShouldBeTrue)
			So(claims.IssuedAt, ShouldEqual, token.IssuedAt().Unix())
			So(claims.Subject, ShouldEqual, "userid1")
			So(claims.ExpiresAt, ShouldEqual, 0)
		})

		Convey("should get a token", func() {
			issuedAt := time.Now()
			claims := jwt.StandardClaims{
				Id:        "tokenid",
				IssuedAt:  issuedAt.Unix(),
				ExpiresAt: issuedAt.Add(time.Hour * 1).Unix(),
				Issuer:    "exampleapp",
				Subject:   "userid1",
			}

			jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedString, err := jwtToken.SignedString([]byte("secret"))
			So(err, ShouldBeNil)

			token := Token{}
			So(store.Get(signedString, &token), ShouldBeNil)

			So(token.UserInfoID, ShouldEqual, "userid1")
			So(token.IssuedAt().Unix(), ShouldEqual, issuedAt.Unix())
			So(token.ExpiredAt.Unix(), ShouldEqual, issuedAt.Add(time.Hour*1).Unix())
		})
	})
}
