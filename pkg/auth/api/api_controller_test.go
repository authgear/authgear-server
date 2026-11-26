package api

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestAPIControllerLoginLink(t *testing.T) {
	Convey("APIController login link", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		loginLinks := NewMockAPIControllerLoginLinkService(ctrl)
		c := &APIController{LoginLinks: loginLinks}

		Convey("passes the code to the service", func() {
			loginLinks.EXPECT().Inspect(gomock.Any(), "CODE").Return(nil)
			So(c.InspectLoginLink(ctx, map[string]any{"code": "CODE"}), ShouldBeNil)

			loginLinks.EXPECT().Approve(gomock.Any(), "CODE").Return(nil)
			So(c.ApproveLoginLink(ctx, map[string]any{"code": "CODE"}), ShouldBeNil)
		})

		Convey("returns the service error", func() {
			serviceErr := errors.New("service error")
			loginLinks.EXPECT().Approve(gomock.Any(), "CODE").Return(serviceErr)

			So(c.ApproveLoginLink(ctx, map[string]any{"code": "CODE"}), ShouldBeError, serviceErr)
		})

		Convey("rejects an invalid request without calling the service", func() {
			for _, body := range []map[string]any{
				nil,
				{},
				{"code": ""},
				{"code": 1},
				{"code": "CODE", "other": 1},
			} {
				err := c.InspectLoginLink(ctx, body)
				So(apierrors.AsAPIErrorWithContext(ctx, err).Reason, ShouldEqual, "ValidationFailed")

				err = c.ApproveLoginLink(ctx, body)
				So(apierrors.AsAPIErrorWithContext(ctx, err).Reason, ShouldEqual, "ValidationFailed")
			}
		})
	})
}
