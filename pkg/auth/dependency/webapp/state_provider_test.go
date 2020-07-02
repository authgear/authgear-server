package webapp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/log"
)

func TestStateProvider(t *testing.T) {
	Convey("StateProviderImpl", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := NewMockStateStore(ctrl)
		p := StateProviderImpl{
			StateStore: store,
			Logger:     StateProviderLogger{log.Null},
		}

		Convey("UpdateError panics if id is invalid", func() {
			store.EXPECT().Get(gomock.Eq("a")).Return(nil, ErrStateNotFound)
			So(func() { p.UpdateError("a", errors.New("error")) }, ShouldPanic)
		})

		Convey("UpdateError does not panic if id is valid", func() {
			s := &State{}
			store.EXPECT().Get(gomock.Eq("a")).Return(s, nil)
			store.EXPECT().Set(gomock.Eq(s)).Return(nil)

			So(func() { p.UpdateError("a", errors.New("error")) }, ShouldNotPanic)
		})

		Convey("CreateState ignore existing sid", func() {
			r, _ := http.NewRequest("GET", "/?x_sid=a", nil)
			_ = r.ParseForm()
			store.EXPECT().Set(gomock.Any()).Return(nil)

			p.CreateState(r, nil)
			So(r.URL.Query().Get("x_sid"), ShouldNotEqual, "a")
		})

		Convey("UpdateState respect existing sid", func() {
			r, _ := http.NewRequest("GET", "/?x_sid=a", nil)
			s := &State{}
			store.EXPECT().Get(gomock.Eq("a")).Return(s, nil)
			store.EXPECT().Set(gomock.Eq(s)).Return(nil)

			So(func() { p.UpdateState(r, nil) }, ShouldNotPanic)
		})

		Convey("UpdateState reject missing sid", func() {
			r, _ := http.NewRequest("GET", "/", nil)
			store.EXPECT().Get(gomock.Eq("")).Return(nil, ErrStateNotFound)

			So(func() { p.UpdateState(r, nil) }, ShouldPanic)
		})

		Convey("RestoreState reject missing sid", func() {
			r, _ := http.NewRequest("GET", "/", nil)
			store.EXPECT().Get(gomock.Eq("")).Return(nil, ErrStateNotFound)

			s, err := p.RestoreState(r, false)
			So(s, ShouldBeNil)
			So(err, ShouldEqual, ErrStateNotFound)
		})

		Convey("RestoreState allow missing sid", func() {
			r, _ := http.NewRequest("GET", "/", nil)

			s, err := p.RestoreState(r, true)
			So(s, ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("RestoreState reject invalid sid", func() {
			r, _ := http.NewRequest("GET", "/?x_sid=a", nil)
			store.EXPECT().Get(gomock.Eq("a")).Return(nil, ErrStateNotFound)

			s, err := p.RestoreState(r, false)
			So(s, ShouldBeNil)
			So(err, ShouldEqual, ErrStateNotFound)
		})

		Convey("RestoreState restores r.Form", func() {
			state := &State{
				Form:  "x_sid=a&b=42",
				Error: &skyerr.APIError{},
			}
			r, _ := http.NewRequest("GET", "/?x_sid=a", nil)
			_ = r.ParseForm()

			store.EXPECT().Get(gomock.Eq("a")).Return(state, nil)

			s, err := p.RestoreState(r, false)
			So(s, ShouldEqual, state)
			So(err, ShouldBeNil)
			So(r.Form.Get("b"), ShouldEqual, "42")
			So(s.Error, ShouldNotBeNil)
		})

		Convey("RestoreState clears error", func() {
			state := &State{
				Form:  "x_sid=a&b=42",
				Error: &skyerr.APIError{},
			}
			r, _ := http.NewRequest("GET", "/?x_sid=a&b=24", nil)
			_ = r.ParseForm()

			store.EXPECT().Get(gomock.Eq("a")).Return(state, nil)
			store.EXPECT().Set(gomock.Eq(state)).Return(nil)

			s, err := p.RestoreState(r, false)
			So(s, ShouldEqual, state)
			So(err, ShouldBeNil)
			So(r.Form.Get("b"), ShouldEqual, "24")
			So(s.Error, ShouldBeNil)
		})
	})
}
