package handler

import (
	"testing"

	"github.com/oursky/skygear/router"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHomeHandler(t *testing.T) {
	Convey("HomeHandler", t, func() {
		req := router.Payload{}
		resp := router.Response{}

		handler := &HomeHandler{}
		handler.Handle(&req, &resp)
		So(resp.Result, ShouldHaveSameTypeAs, statusResponse{})
		s := resp.Result.(statusResponse)
		So(s.Status, ShouldEqual, "OK")
	})
}
