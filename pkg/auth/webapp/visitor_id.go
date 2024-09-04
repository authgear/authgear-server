package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var visitorIDCookieMaxAge = int(duration.VisitorIDLifeTime.Seconds())

var VisitorIDCookieDef = &httputil.CookieDef{
	NameSuffix:        "visitor_id",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteNoneMode, // Ensure it can be read after redirecting from OAuth providers
	MaxAge:            &visitorIDCookieMaxAge,
	IsNonHostOnly:     false,
}

type visitorIDContextKeyType struct{}

var visitorIDContextKey = visitorIDContextKeyType{}

type visitorIDContext struct {
	VisitorID string
}

func WithVisitorID(ctx context.Context, visitorID string) context.Context {
	v, ok := ctx.Value(visitorIDContextKey).(*visitorIDContext)
	if ok {
		v.VisitorID = visitorID
		return ctx
	}

	return context.WithValue(ctx, visitorIDContextKey, &visitorIDContext{
		VisitorID: visitorID,
	})
}

func GetVisitorID(ctx context.Context) string {
	v, ok := ctx.Value(visitorIDContextKey).(*visitorIDContext)
	if !ok {
		return ""
	}
	return v.VisitorID
}
