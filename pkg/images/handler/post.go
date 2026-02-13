package handler

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	gohttputil "net/http/httputil"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	imagesservice "github.com/authgear/authgear-server/pkg/images/service"
	"github.com/authgear/authgear-server/pkg/lib/images"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigurePostRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/_images/:appid/:objectid")
}

type PresignProvider interface {
	Verify(r *http.Request) error
}

type ImagesStore interface {
	Create(ctx context.Context, file *images.File) error
}

var PostHandlerLogger = slogutil.NewLogger("post-handler")

type PostHandlerCloudStorageService interface {
	PresignPutRequest(ctx context.Context, r *imagesservice.PresignUploadRequest) (*imagesservice.PresignUploadResponse, error)
}

type PostHandler struct {
	PostHandlerCloudStorageService PostHandlerCloudStorageService
	PresignProvider                PresignProvider
	Database                       *appdb.Handle
	ImagesStore                    ImagesStore
	Clock                          clock.Clock
}

// nolint:gocognit
func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := r.Context()
	logger := PostHandlerLogger.GetLogger(ctx)

	defer func() {
		if err != nil {
			if !apierrors.IsAPIError(err) {
				logger.WithError(err).Error(ctx, "failed to upload image")
			}
			httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		}
	}()

	err = h.PresignProvider.Verify(r)
	if err != nil {
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		err = apierrors.NewInvalid("invalid content-type")
		return
	}
	if mediaType != "multipart/form-data" {
		err = apierrors.NewInvalid("invalid content-type")
		return
	}
	boundary := params["boundary"]
	if boundary == "" {
		err = apierrors.NewInvalid("invalid boundary")
		return
	}

	reader := multipart.NewReader(r.Body, boundary)

	// At most 5MiB in memory.
	form, err := reader.ReadForm(5 * 1024 * 1024)
	if err != nil {
		err = apierrors.NewInvalid("failed to read request body")
		return
	}
	defer func() {
		if err := form.RemoveAll(); err != nil {
			logger.WithError(err).Error(ctx, "failed to run form remove all")
		}
	}()

	key := ExtractKey(r)
	// Transform the form into PresignUploadRequest.
	presignUploadRequest := imagesservice.PresignUploadRequest{
		Key:     key,
		Headers: map[string]interface{}{},
	}

	// Transform the file field.
	var fileHeader *multipart.FileHeader
	if len(form.File) != 1 {
		err = apierrors.NewInvalid("expected exactly 1 file part")
		return
	}
	for fileFieldName, fileHeaders := range form.File {
		if fileFieldName != "file" {
			err = apierrors.NewInvalid("invalid file field")
			return
		}
		if len(fileHeaders) != 1 {
			err = apierrors.NewInvalid("invalid file field")
			return
		}
		fileHeader = fileHeaders[0]
		presignUploadRequest.Headers["content-length"] = strconv.FormatInt(fileHeader.Size, 10)
		// Only set content-type if content-type does not appear in the form.
		if _, ok := presignUploadRequest.Headers["content-type"]; !ok {
			fileContentType := fileHeader.Header.Get("Content-Type")
			if fileContentType != "" {
				presignUploadRequest.Headers["content-type"] = fileContentType
			}
		}
	}

	encodedMetaDate := r.URL.Query().Get(images.QueryMetadata)
	metadata, err := images.DecodeFileMetadata(ctx, encodedMetaDate)
	if err != nil {
		return
	}
	saveImagesFileRecord := func(ctx context.Context) error {
		objectID := httproute.GetParam(r, "objectid")
		return h.Database.WithTx(ctx, func(ctx context.Context) error {
			return h.ImagesStore.Create(ctx, &images.File{
				ID:        objectID,
				Metadata:  metadata,
				Size:      fileHeader.Size,
				CreatedAt: h.Clock.NowUTC(),
			})
		})
	}

	presignUploadResponse, err := h.PostHandlerCloudStorageService.PresignPutRequest(ctx, &presignUploadRequest)
	if err != nil {
		return
	}

	clientBody, err := fileHeader.Open()
	if err != nil {
		err = fmt.Errorf("failed to open file in form: %w", err)
		return
	}

	director := func(req *http.Request) {
		req.Method = presignUploadResponse.Method
		u, _ := url.Parse(presignUploadResponse.URL)
		req.URL = u
		req.Host = ""
		req.Header.Set("Host", u.Hostname())
		req.ContentLength = fileHeader.Size
		req.Header = http.Header{}
		for _, headerField := range presignUploadResponse.Headers {
			req.Header.Add(headerField.Name, headerField.Value)
		}
		req.Body = clientBody
	}

	modifyResponse := func(resp *http.Response) error {
		// We only know how to modify 2xx response.
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil
		}

		err := saveImagesFileRecord(ctx)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to save image file record")
			return err
		}

		resp.StatusCode = 200
		body := &api.Response{Result: map[string]interface{}{
			"url": fmt.Sprintf("authgearimages:///%s", key),
		}}
		bodyBytes, err := body.EncodeToJSON(ctx)
		if err != nil {
			return err
		}
		resp.ContentLength = int64(len(bodyBytes))
		resp.Header = http.Header{}
		resp.Header.Set("Content-Type", "application/json")
		resp.Header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		resp.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
		return nil
	}

	reverseProxy := &gohttputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
	}

	reverseProxy.ServeHTTP(w, r)
}
