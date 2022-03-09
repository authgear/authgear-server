package api

import (
	"net/http"
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

func ConfigurePresignImagesUploadRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/images/upload")
}

type PresignImagesUploadResponse struct {
	UploadURL string `json:"upload_url"`
}

type PresignImagesUploadHandler struct {
	JSON             JSONResponseWriter
	HTTPProto        httputil.HTTPProto
	HTTPHost         httputil.HTTPHost
	ImagesUploadHost config.ImagesUploadHost
	AppID            config.AppID
	RateLimiter      RateLimiter
}

func (h *PresignImagesUploadHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// FIXME(images): presigned the url

	userID := session.GetUserID(req.Context())
	err := h.RateLimiter.TakeToken(PresignImagesUploadRateLimitBucket(*userID))
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	host := string(h.HTTPHost)
	if h.ImagesUploadHost != "" {
		host = string(h.ImagesUploadHost)
	}
	u := &url.URL{
		Host:   host,
		Scheme: string(h.HTTPProto),
	}
	u.Path = path.Join(u.Path, "_images", string(h.AppID), uuid.New())

	h.JSON.WriteResponse(resp, &api.Response{Result: &PresignImagesUploadResponse{
		UploadURL: u.String(),
	}})
}
