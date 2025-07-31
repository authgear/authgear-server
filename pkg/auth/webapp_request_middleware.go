package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAppNotFoundHTML = template.RegisterHTML(
	"web/app_not_found.html",
	web.ComponentsHTML...,
)

var WebAppRequestMiddlewareLogger = slogutil.NewLogger("request")

// WebAppRequestMiddleware is placed at /pkg/auth because it depends on /pkg/auth/handler/webapp/viewmodels
// So it CANNOT be replaced at /pkg/auth/webapp.
type WebAppRequestMiddleware struct {
	TrustProxy      config.TrustProxy
	HTTPHost        httputil.HTTPHost
	RootProvider    *deps.RootProvider
	ConfigSource    *configsource.ConfigSource
	TemplateEngine  *template.Engine
	BaseViewModeler *viewmodels.BaseViewModeler
}

func (m *WebAppRequestMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := WebAppRequestMiddlewareLogger.GetLogger(ctx)
		logger.Debug(ctx, "serving request",
			slog.String("request_host", r.Host),
			slog.String("request_url", r.URL.String()),
			slog.String("request_header_host", r.Header.Get("Host")),
			slog.String("request_header_x_forwarded_host", r.Header.Get("X-Forwarded-Host")),
			slog.String("request_header_x_original_host", r.Header.Get("X-Original-Host")),
		)
		err := m.ConfigSource.ProvideContext(ctx, r, func(ctx context.Context, appCtx *config.AppContext) error {
			ctx, ap := m.RootProvider.NewAppProvider(ctx, appCtx)
			ctx = deps.WithAppProvider(ctx, ap)
			otelauthgear.SetProjectID(ctx, string(appCtx.Config.AppConfig.ID))

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return nil
		})
		if err != nil {
			if errors.Is(err, configsource.ErrAppNotFound) {
				data := map[string]interface{}{
					"HTTPHost": string(m.HTTPHost),
				}
				baseViewModel := m.BaseViewModeler.ViewModel(r, w)
				viewmodels.Embed(data, baseViewModel)
				m.TemplateEngine.RenderStatus(w, r, http.StatusNotFound, TemplateWebAppNotFoundHTML, data)
			} else {
				logger.WithError(err).Error(ctx, "failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
	})
}
