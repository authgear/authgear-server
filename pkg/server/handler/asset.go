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
	"path/filepath"
	"strings"

	skyAsset "github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AssetUploadHandler models the handler for asset upload request
type AssetUploadHandler struct {
	AssetStore    skyAsset.Store   `inject:"AssetStore"`
	AccessKey     router.Processor `preprocessor:"accesskey"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

// AssetUploadResponse models the response of asset upload request
type AssetUploadResponse struct {
	PostRequest *skyAsset.PostFileRequest `json:"post-request"`
	Asset       *map[string]interface{}   `json:"asset"`
}

// Setup adds injected pre-processors to preprocessors array
func (h *AssetUploadHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.PluginReady,
	}
}

// GetPreprocessors returns all pre-processors for the handler
func (h *AssetUploadHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle is the handling method of the asset upload request
func (h *AssetUploadHandler) Handle(
	payload *router.Payload,
	response *router.Response,
) {
	filename, ok := payload.Data["filename"].(string)
	if !ok {
		response.Err = skyerr.NewInvalidArgument(
			"Missing filename or filename is invalid",
			[]string{"filename"},
		)
		return
	}

	contentType, ok := payload.Data["content-type"].(string)
	if !ok {
		response.Err = skyerr.NewInvalidArgument(
			"Missing content type or content type is invalid",
			[]string{"content-type"},
		)
		return
	}

	contentSizeFloat, ok := payload.Data["content-size"].(float64)
	if !ok {
		response.Err = skyerr.NewInvalidArgument(
			"Missing content size or content size is invalid",
			[]string{"content-size"},
		)
		return
	}
	contentSize := int64(contentSizeFloat)

	// Add UUID to Filename
	dir, file := filepath.Split(filename)
	file = strings.Join([]string{uuidNew(), file}, "-")
	filename = filepath.Join(dir, file)

	// Generate POST File Request
	assetStore := h.AssetStore
	postRequest, err := assetStore.GeneratePostFileRequest(filename)
	if err != nil {
		response.Err = skyerr.NewError(
			skyerr.UnexpectedError,
			"Fail to generate post file request",
		)
		return
	}

	// Save Asset to DB
	conn := payload.DBConn
	asset := skydb.Asset{
		Name:        filename,
		ContentType: contentType,
		Size:        contentSize,
	}
	if err := conn.SaveAsset(&asset); err != nil {
		response.Err = skyerr.NewResourceSaveFailureErrWithStringID("asset", asset.Name)
		return
	}

	// Add Signer to Asset for Serialization
	if signer, ok := assetStore.(skyAsset.URLSigner); ok {
		asset.Signer = signer
	} else {
		logger := logging.CreateLogger(payload.Context(), "handler")
		logger.Warnf("Failed to acquire asset URLSigner, please check configuration")
		response.Err = skyerr.NewError(skyerr.UnexpectedError, "Failed to sign the url")
		return
	}
	assetMap := skyconv.ToMap((*skyconv.MapAsset)(&asset))

	response.Result = &AssetUploadResponse{
		PostRequest: postRequest,
		Asset:       &assetMap,
	}
}
