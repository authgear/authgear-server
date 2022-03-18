package transport

import (
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/images"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func ConfigurePresignImagesUploadRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/_api/admin/images/upload")
}

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type PresignProvider interface {
	PresignPostRequest(url *url.URL) error
}

type PresignImagesUploadResponse struct {
	UploadURL string `json:"upload_url"`
}

type PresignImagesUploadHandlerLogger struct{ *log.Logger }

func NewPresignImagesUploadHandlerLogger(lf *log.Factory) PresignImagesUploadHandlerLogger {
	return PresignImagesUploadHandlerLogger{lf.New("api-presign-images-upload")}
}

type PresignImagesUploadHandler struct {
	JSON            JSONResponseWriter
	HTTPProto       httputil.HTTPProto
	HTTPHost        httputil.HTTPHost
	AppID           config.AppID
	PresignProvider PresignProvider
	Logger          PresignImagesUploadHandlerLogger
}

func (h *PresignImagesUploadHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	metadata := &images.FileMetadata{
		UploadedBy: images.UploadedByTypeAdminAPI,
	}
	encodedData, err := images.EncodeFileMetaData(metadata)
	if err != nil {
		h.Logger.WithError(err).Error("failed to encode metadata")
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	host := string(h.HTTPHost)
	u := &url.URL{
		Host:   host,
		Scheme: string(h.HTTPProto),
	}
	u.Path = path.Join("/_images", string(h.AppID), uuid.New())
	q := u.Query()
	q.Set(images.QueryMetadata, encodedData)
	u.RawQuery = q.Encode()

	err = h.PresignProvider.PresignPostRequest(u)
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	h.JSON.WriteResponse(resp, &api.Response{Result: &PresignImagesUploadResponse{
		UploadURL: u.String(),
	}})
}
