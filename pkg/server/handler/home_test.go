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

package handler

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/router"
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
