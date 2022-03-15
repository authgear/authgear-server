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

type PresignProvider interface {
	PresignPostRequest(url *url.URL) error
}

type PresignImagesUploadResponse struct {
	UploadURL string `json:"upload_url"`
}

type PresignImagesUploadHandler struct {
	JSON            JSONResponseWriter
	HTTPProto       httputil.HTTPProto
	HTTPHost        httputil.HTTPHost
	AppID           config.AppID
	RateLimiter     RateLimiter
	PresignProvider PresignProvider
}

func (h *PresignImagesUploadHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	userID := session.GetUserID(req.Context())
	err := h.RateLimiter.TakeToken(PresignImagesUploadRateLimitBucket(*userID))
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	host := string(h.HTTPHost)
	u := &url.URL{
		Host:   host,
		Scheme: string(h.HTTPProto),
	}
	u.Path = path.Join("/_images", string(h.AppID), uuid.New())

	err = h.PresignProvider.PresignPostRequest(u)
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	h.JSON.WriteResponse(resp, &api.Response{Result: &PresignImagesUploadResponse{
		UploadURL: u.String(),
	}})
}
