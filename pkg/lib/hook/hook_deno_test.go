package hook

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type withoutCancelMatcher struct{}

func (withoutCancelMatcher) Matches(x interface{}) bool {
	if ctx, ok := x.(context.Context); ok {
		s := fmt.Sprintf("%v", ctx)
		if strings.HasSuffix(s, "WithoutCancel") {
			return true
		}
	}
	return false
}

func (withoutCancelMatcher) String() string {
	return "matches context.WithoutContext"
}

func TestDenoHook(t *testing.T) {
	Convey("EventDenoHook", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		syncDenoClient := NewMockSyncDenoClient(ctrl)
		asyncDenoClient := NewMockAsyncDenoClient(ctrl)
		resourceManager := NewMockResourceManager(ctrl)
		denohook := &EventDenoHookImpl{
			DenoHook:        DenoHook{ResourceManager: resourceManager},
			AsyncDenoClient: asyncDenoClient,
			SyncDenoClient:  syncDenoClient,
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
			syncDenoClient.EXPECT().Run(gomock.Any(), "script", e).Times(1).Return(map[string]interface{}{
				"is_allowed": true,
			}, nil)

			ctx := context.Background()
			actual, err := denohook.DeliverBlockingEvent(ctx, u, e)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, resp)
		})

		Convey("DeliverNonBlockingEvent", func() {
			e := &event.Event{}
			u, _ := url.Parse("authgeardeno:///deno/a.ts")

			resourceManager.EXPECT().Read(DenoFile, resource.AppFile{
				Path: "deno/a.ts",
			}).Times(1).Return([]byte("script"), nil)
			asyncDenoClient.EXPECT().Run(withoutCancelMatcher{}, "script", e).Times(1).Return(nil, nil)

			ctx := context.Background()
			err := denohook.DeliverNonBlockingEvent(ctx, u, e)
			runtime.Gosched()
			time.Sleep(500 * time.Millisecond)
			So(err, ShouldBeNil)
		})
	})
}
