package http

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemoveSkygearHeader(t *testing.T) {
	Convey("RemoveSkygearHeader", t, func() {
		So(RemoveSkygearHeader(http.Header{
			"x-skygear-a": {},
			"X-Skygear-B": {},
			"X-SKYGEAR-C": {},
			"x-skygeaR-D": {},
			"host":        {},
			"Accept":      {},
		}), ShouldResemble, http.Header{
			"host":   {},
			"Accept": {},
		})
	})
}
