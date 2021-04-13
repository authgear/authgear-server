package oidc

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type FakeSession struct {
	ID   string
	Type session.Type
}

func (s *FakeSession) SessionID() string {
	return s.ID
}

func (s *FakeSession) SessionType() session.Type {
	return s.Type
}

func TestSID(t *testing.T) {
	Convey("EncodeSID and DecodeSID", t, func() {
		s := &FakeSession{
			ID:   "a",
			Type: session.TypeIdentityProvider,
		}
		typ, sessionID, ok := DecodeSID(EncodeSID(s))
		So(typ, ShouldEqual, session.TypeIdentityProvider)
		So(sessionID, ShouldEqual, "a")
		So(ok, ShouldBeTrue)

		s = &FakeSession{
			ID:   "b",
			Type: session.TypeOfflineGrant,
		}
		typ, sessionID, ok = DecodeSID(EncodeSID(s))
		So(typ, ShouldEqual, session.TypeOfflineGrant)
		So(sessionID, ShouldEqual, "b")
		So(ok, ShouldBeTrue)

		s = &FakeSession{
			ID:   "c",
			Type: "nonsense",
		}
		typ, sessionID, ok = DecodeSID(EncodeSID(s))
		So(ok, ShouldBeFalse)
	})
}
