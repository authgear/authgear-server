package flows

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
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

		Convey("UpdateState panic if state is from RestoreReadOnlyState", func() {
			store.EXPECT().GetState(gomock.Eq("a")).DoAndReturn(func(_ string) (*State, error) {
				return NewState(), nil
			})

			r, _ := http.NewRequest("GET", "?x_sid=a", nil)
			s, err := p.RestoreReadOnlyState(r, false)
			So(err, ShouldBeNil)

			So(func() { p.UpdateState(s, nil, nil) }, ShouldPanic)
		})

		Convey("CloneState generates new InstanceID", func() {
			store.EXPECT().GetState(gomock.Eq("a")).DoAndReturn(func(_ string) (*State, error) {
				state := NewState()
				state.InstanceID = "a"
				return state, nil
			})

			r, _ := http.NewRequest("GET", "?x_sid=a", nil)
			s, err := p.CloneState(r)
			So(err, ShouldBeNil)
			So(s.InstanceID, ShouldNotEqual, "a")
		})

		Convey("CloneState reset Error to nil", func() {
			store.EXPECT().GetState(gomock.Eq("a")).DoAndReturn(func(_ string) (*State, error) {
				state := NewState()
				state.Error = &skyerr.APIError{}
				return state, nil
			})

			r, _ := http.NewRequest("GET", "?x_sid=a", nil)
			s, err := p.CloneState(r)
			So(err, ShouldBeNil)
			So(s.Error, ShouldBeNil)
		})
	})
}
