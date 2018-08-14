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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	skyAsset "github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// used to clean file path
var sanitizedPathRe = regexp.MustCompile(`\A[/.]+`)

func clean(p string) string {
	sanitized := strings.Replace(sanitizedPathRe.ReplaceAllString(path.Clean(p), ""), "..", "", -1)
	// refs #426: S3 Asset Store is not able to put filename with `+` correctly
	sanitized = strings.Replace(sanitized, "+", "", -1)
	return sanitized
}

func validateAssetGetRequest(assetStore skyAsset.Store, fileName string, expiredAtUnix int64, signature string) skyerr.Error {
	// check whether the request is expired
	expiredAt := time.Unix(expiredAtUnix, 0)
	if timeNow().After(expiredAt) {
		return skyerr.NewError(skyerr.PermissionDenied, "Access denied")
	}

	// check the signature of the URL
	signatureParser := assetStore.(skyAsset.SignatureParser)
	valid, err := signatureParser.ParseSignature(signature, fileName, expiredAt)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to parse signature")

		return skyerr.NewError(skyerr.PermissionDenied, "Access denied")
	}

	if !valid {
		return skyerr.NewError(skyerr.InvalidSignature, "Invalid signature")
	}
	return nil
}

func parseRangeHeader(rangeHeader string) (skyAsset.FileRange, error) {
	splits := strings.SplitN(rangeHeader, "=", 2)
	if len(splits) != 2 {
		return skyAsset.FileRange{}, errors.New("range header is malformed")
	}

	if strings.ToLower(splits[0]) != "bytes" {
		return skyAsset.FileRange{}, errors.New(
			"only support range in unit of bytes",
		)
	}

	rangeSplits := strings.SplitN(splits[1], "-", 2)
	if len(rangeSplits) != 2 {
		return skyAsset.FileRange{}, errors.New("the byte range is malformed")
	}

	rangeFrom, err1 := strconv.ParseInt(rangeSplits[0], 10, 64)
	rangeTo, err2 := strconv.ParseInt(rangeSplits[1], 10, 64)
	if err1 != nil || err2 != nil {
		return skyAsset.FileRange{}, errors.New("the byte range is malformed")
	}

	if rangeTo < rangeFrom {
		rangeTo = rangeFrom
	}

	return skyAsset.FileRange{
		From: rangeFrom,
		To:   rangeTo,
	}, nil
}

// GetFileHandler models the handler for getting asset file
type GetFileHandler struct {
	AssetStore    skyAsset.Store   `inject:"AssetStore"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

// Setup sets preprocessors being used
func (h *GetFileHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DBConn,
	}
}

// GetPreprocessors returns all preprocessors
func (h *GetFileHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle handles the get request for asset file
func (h *GetFileHandler) Handle(payload *router.Payload, response *router.Response) {
	logger := logging.CreateLogger(payload.Context(), "handler")
	payload.Req.ParseForm()

	store := h.AssetStore
	fileName := clean(payload.Params[0])
	if store.(skyAsset.URLSigner).IsSignatureRequired() {
		expiredAtUnix, err := strconv.ParseInt(payload.Req.Form.Get("expiredAt"), 10, 64)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect expiredAt to be an integer")
			return
		}

		signature := payload.Req.Form.Get("signature")
		requestErr := validateAssetGetRequest(h.AssetStore, fileName, expiredAtUnix, signature)
		if requestErr != nil {
			response.Err = requestErr
			return
		}
	}

	// everything's right, proceed with the request
	asset := &skydb.Asset{}
	if err := payload.DBConn.GetAsset(fileName, asset); err != nil {
		logger.WithError(err).Errorf("Failed to get asset")

		response.Err = skyerr.NewResourceFetchFailureErr("asset", fileName)
		return
	}

	rangeHeader := payload.Req.Header.Get("Range")
	if rangeHeader != "" {
		byteRange, err := parseRangeHeader(payload.Req.Header.Get("Range"))
		if err == nil {
			h.handlePartialRangedRequest(asset, byteRange, payload, response, logger)
			return
		}

		logger.WithError(err).Error("Error in parsing range header")
	}

	h.handleFullRangedRequest(asset, payload, response, logger)
}

func (h *GetFileHandler) handlePartialRangedRequest(
	asset *skydb.Asset,
	byteRange skyAsset.FileRange,
	payload *router.Payload,
	response *router.Response,
	logger *logrus.Entry,
) {
	store := h.AssetStore
	fileName := asset.Name

	fileRangedGetter, supported := store.(skyAsset.FileRangedGetter)
	if !supported {
		response.Err = skyerr.NewError(
			skyerr.NotSupported,
			"Getting asset with a byte range is supported",
		)
		return
	}

	writer := response.Writer()
	result, err := fileRangedGetter.GetRangedFileReader(fileName, byteRange)
	if err != nil {
		notAcceptedError, isNotAccepted := err.(skyAsset.FileRangeNotAcceptedError)
		if isNotAccepted {
			writer.Header().Set("Content-Type", "text/plain")
			writer.Header().Set("Content-Length", "0")
			writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)

			writer.Write([]byte(notAcceptedError.Error()))

			return
		}

		logger.WithError(err).Error("Error when getting ranged file reader")
		response.Err = skyerr.NewResourceFetchFailureErr("asset", fileName)
		return
	}

	readCloser := result.ReadCloser
	defer readCloser.Close()

	contentLength := result.AcceptedRange.To - result.AcceptedRange.From + 1
	contentRange := fmt.Sprintf(
		"bytes %d-%d/%d",
		result.AcceptedRange.From,
		result.AcceptedRange.To,
		result.TotalSize,
	)

	writer.Header().Set("Content-Type", asset.ContentType)
	writer.Header().Set("Content-Range", contentRange)
	writer.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	writer.WriteHeader(http.StatusPartialContent)

	if _, err := io.CopyN(writer, readCloser, contentLength); err != nil {
		// there is nothing we can do if error occurred after started
		// writing a response. Log.
		logger.WithError(err).Errorf("Error writing file to response")
	}
}

func (h *GetFileHandler) handleFullRangedRequest(
	asset *skydb.Asset,
	payload *router.Payload,
	response *router.Response,
	logger *logrus.Entry,
) {
	store := h.AssetStore
	fileName := asset.Name
	reader, err := store.GetFileReader(fileName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get file reader")

		response.Err = skyerr.NewResourceFetchFailureErr("asset", fileName)
		return
	}
	defer reader.Close()

	writer := response.Writer()
	if writer == nil {
		// The response is already written.
		return
	}

	writer.Header().Set("Content-Type", asset.ContentType)
	writer.Header().Set("Content-Length", strconv.FormatInt(asset.Size, 10))

	if _, err := io.Copy(writer, reader); err != nil {
		// there is nothing we can do if error occurred after started
		// writing a response. Log.
		logger.WithError(err).Errorf("Error writing file to response")
	}
}

// UploadFileHandler receives and persists a file to be associated by Record.
//
// Example curl (PUT):
//	curl -XPUT \
//		-H 'X-Skygear-API-Key: apiKey' \
//		-H 'Content-Type: text/plain' \
//		--data-binary '@file.txt' \
//		http://localhost:3000/files/filename
//
// Example curl (POST):
//	curl -XPOST \
//    -H "X-Skygear-API-Key: apiKey" \
//    -F 'file=@file.txt' \
//    http://localhost:3000/files/filename
//
type UploadFileHandler struct {
	AssetStore    skyAsset.Store   `inject:"AssetStore"`
	AccessKey     router.Processor `preprocessor:"accesskey"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

type uploadFileRequest struct {
	filename    string
	contentType string
	fileReader  io.Reader
}

// Setup sets preprocessors being used
func (h *UploadFileHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
	}
}

// GetPreprocessors returns all preprocessors
func (h *UploadFileHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle handles the upload asset request
func (h *UploadFileHandler) Handle(
	payload *router.Payload,
	response *router.Response,
) {

	logger := logging.CreateLogger(payload.Context(), "handler")
	uploadRequest, err := parseUploadFileRequest(payload)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}

	if uploadRequest.contentType == "" {
		response.Err = skyerr.NewError(
			skyerr.InvalidArgument,
			"Content-Type cannot be empty",
		)
		return
	}

	written, tempFile, err := copyToTempFile(uploadRequest.fileReader)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	if written == 0 {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "Zero-byte content")
		return
	}

	asset := skydb.Asset{}
	conn := payload.DBConn
	if err := conn.GetAsset(uploadRequest.filename, &asset); err != nil {
		// compatible with SDK <= v0.15
		dir, file := filepath.Split(uploadRequest.filename)
		file = strings.Join([]string{uuidNew(), file}, "-")

		asset.Name = filepath.Join(dir, file)
		asset.ContentType = uploadRequest.contentType
	}

	assetStore := h.AssetStore
	if err := assetStore.PutFileReader(
		asset.Name,
		tempFile,
		written,
		asset.ContentType,
	); err != nil {

		response.Err = skyerr.MakeError(err)
		return
	}

	asset.Size = written
	if err := conn.SaveAsset(&asset); err != nil {
		response.Err = skyerr.NewResourceSaveFailureErrWithStringID("asset", asset.Name)
		return
	}

	if signer, ok := h.AssetStore.(skyAsset.URLSigner); ok {
		asset.Signer = signer
	} else {
		logger.Warnf("Failed to acquire asset URLSigner, please check configuration")
		response.Err = skyerr.NewError(skyerr.UnexpectedError, "Failed to sign the url")
		return
	}
	response.Result = skyconv.ToMap((*skyconv.MapAsset)(&asset))
}

// parseUploadFileRequest tries to parse the payload from router to be compatible
// with both PUT requests and multiparts POST request
func parseUploadFileRequest(payload *router.Payload) (*uploadFileRequest, error) {
	logger := logging.CreateLogger(payload.Context(), "handler")
	httpRequest := payload.Req
	method := httpRequest.Method

	var (
		filename, contentType string
		fileReader            io.ReadCloser
	)

	if method == http.MethodPost {
		// use 100 MB max memory to parse the multiparts Form
		err := httpRequest.ParseMultipartForm(100 << 20)
		if err != nil {
			logger.
				WithError(err).
				Error("Fail to parse multiparts form for asset upload")

			return nil, err
		}

		form := httpRequest.MultipartForm
		fileHeader := form.File["file"]
		if fileHeader == nil || len(fileHeader) == 0 {
			logger.Error("Missing file in multiparts form")

			return nil, errors.New("Missing file in multiparts form")
		}

		firstFileHeader := fileHeader[0]

		filename = clean(payload.Params[0])
		contentType = firstFileHeader.Header["Content-Type"][0]
		fileReader, err = firstFileHeader.Open()
		if err != nil {
			return nil, err
		}
	} else if method == http.MethodPut {
		filename = clean(payload.Params[0])
		contentType = httpRequest.Header.Get("Content-Type")
		fileReader = httpRequest.Body
	} else {
		return nil, errors.New(
			"Method " + method + " is not supported",
		)
	}

	return &uploadFileRequest{
		filename:    filename,
		contentType: contentType,
		fileReader:  fileReader,
	}, nil
}

func copyToTempFile(src io.Reader) (written int64, tempFile *os.File, err error) {
	tempFile, err = ioutil.TempFile("", "")
	if err != nil {
		return
	}
	written, err = io.Copy(tempFile, src)
	if err != nil {
		cleanupFile(tempFile)
		tempFile = nil
		return
	}
	if _, err = tempFile.Seek(0, 0); err != nil {
		cleanupFile(tempFile)
		tempFile = nil
		return
	}
	return
}

func cleanupFile(f *os.File) error {
	closeErr := f.Close()
	if closeErr != nil {
		logrus.WithError(closeErr).Errorf("Failed to close tempFile %s", f.Name())
		return closeErr
	}

	if err := os.Remove(f.Name()); err != nil {
		logrus.WithError(err).Errorf("Failed to remove file %s: %v", f.Name())
		return err
	}

	return nil
}
