// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package preprocessor

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skyerr"
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
