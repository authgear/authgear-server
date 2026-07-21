package session_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSessionCookieDef(t *testing.T) {
	Convey("NewSessionCookieDef", t, func() {
		Convey("is a persistent cookie by default", func() {
			cfg := &config.SessionConfig{Lifetime: 3600}
			def := session.NewSessionCookieDef(cfg)

			So(def.Def.MaxAge, ShouldNotBeNil)
			So(*def.Def.MaxAge, ShouldEqual, 3600)
			So(def.SameSiteStrictDef.MaxAge, ShouldNotBeNil)
			So(*def.SameSiteStrictDef.MaxAge, ShouldEqual, 3600)
		})

		Convey("is a HTTP session cookie when UseSessionCookie is true", func() {
			cfg := &config.SessionConfig{Lifetime: 3600, UseSessionCookie: true}
			def := session.NewSessionCookieDef(cfg)

			So(def.Def.MaxAge, ShouldBeNil)
			So(def.SameSiteStrictDef.MaxAge, ShouldBeNil)
		})
	})
}
