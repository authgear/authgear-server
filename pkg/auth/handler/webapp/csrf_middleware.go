package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var CSRFMiddlewareLogger = slogutil.NewLogger("webapp-csrf-middleware")

type CSRFMiddlewareUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type CSRFMiddleware struct {
	TrustProxy              config.TrustProxy
	BaseViewModel           *viewmodels.BaseViewModeler
	Renderer                Renderer
	UIImplementationService CSRFMiddlewareUIImplementationService
	EnvironmentConfig       *config.EnvironmentConfig
}

func (m *CSRFMiddleware) Handle(next http.Handler) http.Handler {
	if m.EnvironmentConfig.End2EndCSRFProtectionDisabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := CSRFMiddlewareLogger.GetLogger(ctx)
		secFetchError := httputil.NewCrossOriginProtection().Check(r)
		if secFetchError != nil {
			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterSecFetchCSRFRequestCount,
				otelauthgear.WithStatusError(),
			)
			logger.WithError(secFetchError).Warn(ctx, "SecFetch CSRF Forbidden")

			uiImpl := m.UIImplementationService.GetUIImplementation()

			data := make(map[string]interface{})
			baseViewModel := m.BaseViewModel.ViewModelForAuthFlow(r, w)
			viewmodels.Embed(data, baseViewModel)

			switch uiImpl {
			case config.UIImplementationInteraction:
				http.Error(w, fmt.Sprintf("%v handler/auth/webapp",
					http.StatusText(http.StatusForbidden)),
					http.StatusForbidden)
			case config.UIImplementationAuthflowV2:
				fallthrough
			default:
				m.Renderer.RenderHTML(w, r, TemplateCSRFErrorHTML, data)
			}
			return
		} else {
			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterSecFetchCSRFRequestCount,
				otelauthgear.WithStatusOk(),
			)
		}

		next.ServeHTTP(w, r)
	})
}
