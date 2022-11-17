package hook

import (
	"context"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
)

func TestDenoHook(t *testing.T) {
	Convey("DenoHook", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		syncDenoClient := NewMockSyncDenoClient(ctrl)
		asyncDenoClient := NewMockAsyncDenoClient(ctrl)
		resourceManager := NewMockResourceManager(ctrl)
		denohook := &DenoHookImpl{
			Context:         ctx,
			SyncDenoClient:  syncDenoClient,
			AsyncDenoClient: asyncDenoClient,
			ResourceManager: resourceManager,
		}

		Convey("DeliverBlockingEvent", func() {
			e := &event.Event{}
			u, _ := url.Parse("http://localhost/")
			resp := &event.HookResponse{
				IsAllowed: true,
			}

			resourceManager.EXPECT().Read(gomock.Any(), gomock.Any()).Times(1).Return([]byte(nil), nil)
			syncDenoClient.EXPECT().Run(ctx, "", e).Times(1).Return(map[string]interface{}{
				"is_allowed": true,
			}, nil)

			actual, err := denohook.DeliverBlockingEvent(u, e)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, resp)
		})

		Convey("DeliverNonBlockingEvent", func() {
			e := &event.Event{}
			u, _ := url.Parse("http://localhost/")

			resourceManager.EXPECT().Read(gomock.Any(), gomock.Any()).Times(1).Return([]byte(nil), nil)
			asyncDenoClient.EXPECT().Run(ctx, "", e).Times(1).Return(nil, nil)

			err := denohook.DeliverNonBlockingEvent(u, e)
			So(err, ShouldBeNil)
		})
	})
}
