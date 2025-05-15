package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type NoProjectCSPMiddleware struct {
	AllowedFrameAncestorsFromEnv config.AllowedFrameAncestors
}

func (m *NoProjectCSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, r := httputil.CSPNoncePerRequest(r)

		var frameAncestors []string

		for _, frameAncestor := range m.AllowedFrameAncestorsFromEnv {
			frameAncestors = append(frameAncestors, frameAncestor)
		}

		cspDirectives, err := web.CSPDirectives(web.CSPDirectivesOptions{
			Nonce:          nonce,
			FrameAncestors: frameAncestors,
		})
		if err != nil {
			panic(err)
		}

		if len(frameAncestors) == 0 {
			w.Header().Set("X-Frame-Options", "DENY")
		}

		w.Header().Set("Content-Security-Policy", cspDirectives.String())
		next.ServeHTTP(w, r)
	})
}
