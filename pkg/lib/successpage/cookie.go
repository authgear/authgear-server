package successpage

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var pathCookieMaxAge = int(duration.Short.Seconds())

var PathCookieDef = &httputil.CookieDef{
	NameSuffix:    "successful_page_path",
	Path:          "/",
	SameSite:      http.SameSiteNoneMode,
	MaxAge:        &pathCookieMaxAge,
	IsNonHostOnly: false,
}
