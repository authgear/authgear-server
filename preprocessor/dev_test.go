package preprocessor

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

func TestDevOnlyProcessor(t *testing.T) {
	Convey("test dev only preprocessor when in dev mode", t, func() {
		pp := DevOnlyProcessor{
			DevMode: true,
		}

		payload := &router.Payload{
			Data: map[string]interface{}{},
			Meta: map[string]interface{}{},
		}
		resp := &router.Response{}

		Convey("test should be okay", func() {
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
		})
	})

	Convey("test dev only preprocessor when not in dev mode", t, func() {
		pp := DevOnlyProcessor{
			DevMode: false,
		}

		payload := &router.Payload{
			Data: map[string]interface{}{},
			Meta: map[string]interface{}{},
		}
		resp := &router.Response{}

		Convey("not okay when client key", func() {
			payload.AccessKey = router.ClientAccessKey
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusForbidden)
			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.PermissionDenied)
		})

		Convey("okay when master key", func() {
			payload.AccessKey = router.MasterAccessKey
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
		})
	})
}
