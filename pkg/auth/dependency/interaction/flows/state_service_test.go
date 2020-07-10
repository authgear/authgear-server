package flows

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/log"
)

func TestStateService(t *testing.T) {
	Convey("StateService", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := NewMockStateStore(ctrl)
		p := StateService{
			ServerConfig: &config.ServerConfig{
				TrustProxy: true,
			},
			StateStore: store,
			Logger:     StateServiceLogger{log.Null},
		}

		Convey("CreateState ignore existing sid", func() {
			r, _ := http.NewRequest("GET", "/?x_sid=a", nil)
			_ = r.ParseForm()
			store.EXPECT().Set(gomock.Any()).Return(nil)

			p.CreateState(r, nil, nil)
			So(r.URL.Query().Get("x_sid"), ShouldNotEqual, "a")
		})

		Convey("UpdateState reject missing sid", func() {
			_, _ = http.NewRequest("GET", "/", nil)

			So(func() { p.UpdateState(nil, nil, nil) }, ShouldPanic)
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
	})
}
