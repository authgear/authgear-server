package api

import (
	"context"
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/images"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type RateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

func ConfigurePresignImagesUploadRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/images/upload")
}

type PresignProvider interface {
	PresignPostRequest(url *url.URL) error
}

type PresignImagesUploadResponse struct {
	UploadURL string `json:"upload_url"`
}

type PresignImagesUploadHandlerLogger struct{ *log.Logger }

type TurboResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, req *http.Request, resp *api.Response)
}

func NewPresignImagesUploadHandlerLogger(lf *log.Factory) PresignImagesUploadHandlerLogger {
	return PresignImagesUploadHandlerLogger{lf.New("api-presign-images-upload")}
}

type PresignImagesUploadHandler struct {
	Turbo           TurboResponseWriter
	HTTPProto       httputil.HTTPProto
	HTTPHost        httputil.HTTPHost
	AppID           config.AppID
	RateLimiter     RateLimiter
	PresignProvider PresignProvider
	Logger          PresignImagesUploadHandlerLogger
}

func (h *PresignImagesUploadHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := session.GetUserID(ctx)
	failed, err := h.RateLimiter.Allow(ctx, PresignImageUploadRequestBucketSpec(*userID))
	if err != nil {
		h.Turbo.WriteResponse(resp, req, &api.Response{Error: err})
		return
	}
	if err := failed.Error(); err != nil {
		h.Turbo.WriteResponse(resp, req, &api.Response{Error: err})
		return
	}

	metadata := &images.FileMetadata{
		UserID:     *userID,
		UploadedBy: images.UploadedByTypeUser,
	}
	encodedData, err := images.EncodeFileMetaData(metadata)
	if err != nil {
		h.Logger.WithError(err).Error("failed to encode metadata")
		h.Turbo.WriteResponse(resp, req, &api.Response{Error: err})
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
		h.Turbo.WriteResponse(resp, req, &api.Response{Error: err})
		return
	}

	h.Turbo.WriteResponse(resp, req, &api.Response{Result: &PresignImagesUploadResponse{
		UploadURL: u.String(),
	}})
}
