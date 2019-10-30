package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/asset/dependency/presign"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	coreIo "github.com/skygeario/skygear-server/pkg/core/io"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var BadAssetUploadForm = skyerr.BadRequest.WithReason("BadAssetUploadForm")

func AttachUploadFormHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/upload_form", &UploadFormHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "POST")
	return server
}

type UploadFormHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *UploadFormHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UploadFormHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h
}

type UploadFormHandler struct {
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
	PresignProvider      presign.Provider      `dependency:"PresignProvider"`
	Validator            *validation.Validator `dependency:"Validator"`
}

func (h *UploadFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	err := h.Handle(w, r)
	if err != nil {
		response.Error = err
		handler.WriteResponse(w, response)
	}
	// If there is no error, the response is written by reverse proxy.
}

func (h *UploadFormHandler) Handle(w http.ResponseWriter, r *http.Request) (err error) {
	// Verify signature
	err = h.PresignProvider.Verify(r)
	if err != nil {
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		err = BadAssetUploadForm.New("invalid content-type")
		return
	}
	if mediaType != "multipart/form-data" {
		err = BadAssetUploadForm.New("invalid content-type")
		return
	}
	boundary := params["boundary"]
	if boundary == "" {
		err = BadAssetUploadForm.New("invalid boundary")
		return
	}

	reader := multipart.NewReader(r.Body, boundary)

	// At most 5MiB in memory.
	form, err := reader.ReadForm(5 * 1024 * 1024)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to read request body")
		return
	}
	defer form.RemoveAll()

	// Transform the form into PresignUploadRequest.
	presignUploadRequest := cloudstorage.PresignUploadRequest{
		Headers: map[string]interface{}{},
	}

	// Transform simple fields.
	for fieldName, values := range form.Value {
		if len(values) != 1 {
			err = BadAssetUploadForm.New(fmt.Sprintf("repeated field: %s", fieldName))
			return
		}
		value := values[0]
		switch fieldName {
		case "prefix":
			presignUploadRequest.Prefix = value
		case "access":
			presignUploadRequest.Access = cloudstorage.AccessType(value)
		default:
			presignUploadRequest.Headers[fieldName] = value
		}
	}

	// Transform the file field.
	var fileHeader *multipart.FileHeader
	if len(form.File) != 1 {
		err = BadAssetUploadForm.New("expected exactly 1 file part")
		return
	}
	for fileFieldName, fileHeaders := range form.File {
		if fileFieldName != "file" {
			err = BadAssetUploadForm.New("invalid file field")
			return
		}
		if len(fileHeaders) != 1 {
			err = BadAssetUploadForm.New("invalid file field")
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

	jsonBytes, err := json.Marshal(presignUploadRequest)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to marshal JSON")
		return
	}
	jsonReader := bytes.NewReader(jsonBytes)

	var validatedPresignUploadRequest cloudstorage.PresignUploadRequest
	err = h.Validator.ParseReader("#PresignUploadRequest", jsonReader, &validatedPresignUploadRequest)
	if err != nil {
		if validationError, ok := err.(validation.Error); ok {
			err = validationError.SkyErrInvalidArgument("Validation Error")
		}
		return
	}
	validatedPresignUploadRequest.SetDefaultValue()

	presignUploadResponse, err := h.CloudStorageProvider.PresignPutRequest(&validatedPresignUploadRequest)
	if err != nil {
		return
	}

	clientBody, err := fileHeader.Open()
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to open file in form")
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

		body := handler.APIResponse{
			Result: map[string]interface{}{
				"asset_name": presignUploadResponse.AssetName,
			},
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}

		resp.ContentLength = int64(len(bodyBytes))
		resp.Header = http.Header{}
		resp.Header.Set("Content-Type", "application/json")
		resp.Header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		resp.Body = &coreIo.BytesReaderCloser{Reader: bytes.NewReader(bodyBytes)}

		return nil
	}

	reverseProxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
	}

	reverseProxy.ServeHTTP(w, r)
	return
}
