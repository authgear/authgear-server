package webapp

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/cimd"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureClientLogoRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "HEAD", "OPTIONS").
		WithPathPattern("/_internals/client_logo")
}

var ClientLogoHandlerLogger = slogutil.NewLogger("client-logo-handler")

type ClientLogoClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

type ClientLogoLogoService interface {
	Get(ctx context.Context, clientID string, logoURI string) (*cimd.LogoResult, error)
}

// ClientLogoHandler proxies a dynamic client's self-asserted logo_uri, so
// the end user's browser only ever talks to Authgear and never to the
// client's own server (spec § Privacy Considerations §9.2). It is not on
// /oauth2/, has no session and no CSRF token -- the consent screen's <img>
// carries neither -- and it must never trigger a metadata document fetch:
// Resolver is a plain read, never cimd.Service.EnsureClientResolved. Note
// there is no cimd.Service field at all here, which is what makes that a
// compile-time guarantee rather than a test.
type ClientLogoHandler struct {
	Resolver ClientLogoClientResolver
	Logos    ClientLogoLogoService
}

func (h *ClientLogoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientID := r.URL.Query().Get("client_id")

	// (1) Resolve the client. A plain read -- Redis, then Postgres -- never
	// a fetch. A client_id with no persisted record is simply a 404 here.
	client := h.Resolver.ResolveClient(ctx, clientID)
	if client == nil || client.LogoURI == "" {
		http.NotFound(rw, r)
		return
	}

	// (2) Dynamic clients only. A static client's logo_uri is admin-chosen
	// and continues to render directly via __brand_logo.html.
	if !client.IsDynamicClient() {
		http.NotFound(rw, r)
		return
	}

	// (3) Serve from cache, or fetch-and-cache.
	result, err := h.Logos.Get(ctx, clientID, client.LogoURI)
	if err != nil {
		// Every logo-specific failure collapses to CIMDLogoUnavailable
		// inside LogoService.Get, so the endpoint cannot report on the
		// reachability of whatever host the document named. An
		// infrastructure failure is NOT that Kind and must not be
		// swallowed as a 404.
		if apierrors.IsKind(err, cimd.CIMDLogoUnavailable) {
			http.NotFound(rw, r)
			return
		}
		logger := ClientLogoHandlerLogger.GetLogger(ctx)
		logger.WithError(err).Error(ctx, "failed to serve client logo")
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", result.ContentType)
	rw.Header().Set("Content-Length", strconv.Itoa(len(result.Body)))
	// Not "no-store": the whole point is that browsers stop re-fetching.
	// "private" because the response is not user-specific but is served on
	// a path that may sit behind a per-tenant CDN; keep it out of shared
	// caches to avoid one project's logo being served for another's host.
	rw.Header().Set("Cache-Control", "private, max-age=3600")
	rw.Header().Set("X-Content-Type-Options", "nosniff")
	// The image is attacker-supplied bytes served from Authgear's own
	// origin. Both headers below are what prevent that from being an XSS
	// vector.
	rw.Header().Set("Content-Disposition", "inline")
	rw.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	http.ServeContent(rw, r, "", result.FetchedAt, bytes.NewReader(result.Body))
}
