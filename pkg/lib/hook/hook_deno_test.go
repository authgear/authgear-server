package hook

import (
	"context"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/util/resource"
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
			u, _ := url.Parse("authgeardeno:///deno/a.ts")
			resp := &event.HookResponse{
				IsAllowed: true,
			}

			resourceManager.EXPECT().Read(DenoFile, resource.AppFile{
				Path: "deno/a.ts",
			}).Times(1).Return([]byte("script"), nil)
			syncDenoClient.EXPECT().Run(ctx, "script", e).Times(1).Return(map[string]interface{}{
				"is_allowed": true,
			}, nil)

			actual, err := denohook.DeliverBlockingEvent(u, e)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, resp)
		})

		Convey("DeliverNonBlockingEvent", func() {
			e := &event.Event{}
			u, _ := url.Parse("authgeardeno:///deno/a.ts")

			resourceManager.EXPECT().Read(DenoFile, resource.AppFile{
				Path: "deno/a.ts",
			}).Times(1).Return([]byte("script"), nil)
			asyncDenoClient.EXPECT().Run(ctx, "script", e).Times(1).Return(nil, nil)

			err := denohook.DeliverNonBlockingEvent(u, e)
			So(err, ShouldBeNil)
		})
	})
}
