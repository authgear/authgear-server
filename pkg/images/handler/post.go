package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/cloudstorage"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigurePostRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST").
		WithPathPattern("/_images/:appid/:objectid")
}

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type PostHandlerLogger struct{ *log.Logger }

func NewPostHandlerLogger(lf *log.Factory) PostHandlerLogger {
	return PostHandlerLogger{lf.New("post-handler")}
}

type PostHandler struct {
	Logger               PostHandlerLogger
	JSON                 JSONResponseWriter
	CloudStorageProvider cloudstorage.Provider
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// FIXME(images): Verify signature
	// FIXME(images): Store records
	// FIXME(images): change images to public

	var err error

	defer func() {
		if err != nil {
			if !apierrors.IsAPIError(err) {
				h.Logger.WithError(err).Error("failed to upload image")
			}
			h.JSON.WriteResponse(w, &api.Response{Error: err})
		}
	}()

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
			h.Logger.WithError(err).Error("failed to run form remove all")
		}
	}()

	key := ExtractKey(r)
	// Transform the form into PresignUploadRequest.
	presignUploadRequest := cloudstorage.PresignUploadRequest{
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

	presignUploadResponse, err := h.CloudStorageProvider.PresignPutRequest(&presignUploadRequest)
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
		resp.StatusCode = 200
		body := &api.Response{Result: map[string]interface{}{
			"url": fmt.Sprintf("authgearimages:///%s", key),
		}}
		bodyBytes, err := json.Marshal(body)
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

	reverseProxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
	}

	reverseProxy.ServeHTTP(w, r)
}
