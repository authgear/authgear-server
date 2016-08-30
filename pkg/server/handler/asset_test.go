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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

type generatePostFileRequestAssetStore struct{}

func (s generatePostFileRequestAssetStore) GetFileReader(
	name string,
) (io.ReadCloser, error) {
	panic("Not Implemented")
}

func (s generatePostFileRequestAssetStore) PutFileReader(
	name string,
	src io.Reader,
	length int64,
	contentType string,
) error {
	panic("Not Implemented")
}

func (s generatePostFileRequestAssetStore) GeneratePostFileRequest(
	name string,
) (*asset.PostFileRequest, error) {
	return &asset.PostFileRequest{
		Action: "http://skygear.dev/files/" + name,
		ExtraFields: map[string]interface{}{
			"X-Extra-Field-1": "extra-fields-1",
			"X-Extra-Field-2": "extra-fields-2",
		},
	}, nil
}

func (s generatePostFileRequestAssetStore) SignedURL(name string) (string, error) {
	return "http://asset.skygear.dev/" + name, nil
}

func (s generatePostFileRequestAssetStore) IsSignatureRequired() bool {
	return false
}

type saveAssetDBConn struct {
	skydb.Conn
	savedAsset map[string]*skydb.Asset
}

func (db *saveAssetDBConn) SaveAsset(asset *skydb.Asset) error {
	db.savedAsset[asset.Name] = asset
	return nil
}

func TestAssetUploadHandler(t *testing.T) {
	Convey("Asset Upload Handler", t, func() {
		assetDBConn := &saveAssetDBConn{}
		assetDBConn.savedAsset = map[string]*skydb.Asset{}

		assetRouter := handlertest.NewSingleRouteRouter(
			&AssetUploadHandler{AssetStore: generatePostFileRequestAssetStore{}},
			func(p *router.Payload) {
				p.DBConn = assetDBConn
			},
		)

		Convey("Success on normal flow", func() {
			uuidNew = func() string {
				return "7b0e2a7c-7135-4912-a6c9-7c1dbec0f5ef"
			}

			res := assetRouter.POST(`{
        "filename": "file001",
        "content-type": "text/plain",
        "content-size": 2384571
      }`)

			So(res.Code, ShouldEqual, http.StatusOK)

			resJSON := struct {
				Result *AssetUploadResponse `json:"result"`
			}{}

			So(json.Unmarshal(res.Body.Bytes(), &resJSON), ShouldBeNil)
			So(resJSON.Result, ShouldNotBeNil)

			So(resJSON.Result.Asset, ShouldNotBeNil)
			So((*resJSON.Result.Asset)["$type"], ShouldEqual, "asset")
			So(
				(*resJSON.Result.Asset)["$name"],
				ShouldEqual,
				"7b0e2a7c-7135-4912-a6c9-7c1dbec0f5ef-file001",
			)
			So(
				(*resJSON.Result.Asset)["$url"],
				ShouldEqual,
				"http://asset.skygear.dev/7b0e2a7c-7135-4912-a6c9-7c1dbec0f5ef-file001",
			)

			So(resJSON.Result.PostRequest, ShouldNotBeNil)
			So(strings.HasSuffix(resJSON.Result.PostRequest.Action, "file001"), ShouldBeTrue)
			So(
				resJSON.Result.PostRequest.ExtraFields["X-Extra-Field-1"],
				ShouldEqual,
				"extra-fields-1",
			)
			So(
				resJSON.Result.PostRequest.ExtraFields["X-Extra-Field-2"],
				ShouldEqual,
				"extra-fields-2",
			)

			savedAsset :=
				assetDBConn.savedAsset["7b0e2a7c-7135-4912-a6c9-7c1dbec0f5ef-file001"]

			So(savedAsset, ShouldNotBeNil)
			So(savedAsset.ContentType, ShouldEqual, "text/plain")
			So(savedAsset.Size, ShouldEqual, 2384571)
		})

		Convey("Fail when no filename", func() {
			res := assetRouter.POST(`{
        "content-type": "text/plain",
        "content-size": 2384571
      }`)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Fail when no content type", func() {
			res := assetRouter.POST(`{
        "filename": "file001",
        "content-size": 2384571
      }`)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Fail when no content-size", func() {
			res := assetRouter.POST(`{
        "filename": "file001",
        "content-type": "text/plain"
      }`)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}
