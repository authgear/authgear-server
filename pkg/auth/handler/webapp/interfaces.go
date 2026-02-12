package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// TODO(tung): Recheck which interface is actually used

type TutorialCookie interface {
	Pop(r *http.Request, rw http.ResponseWriter, name httputil.TutorialCookieName) bool
}

type ErrorService interface {
	HasError(ctx context.Context, r *http.Request) bool
}

type MeterService interface {
	TrackPageView(ctx context.Context, visitorID string, pageType meter.PageType) error
}
