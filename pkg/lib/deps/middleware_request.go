package deps

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAppNotFoundHTML = template.RegisterHTML(
	"web/app_not_found.html",
	web.ComponentsHTML...,
)

type RequestMiddleware struct {
	HTTPHost        httputil.HTTPHost
	RootProvider    *RootProvider
	ConfigSource    *configsource.ConfigSource
	TemplateEngine  *template.Engine
	BaseViewModeler *viewmodels.BaseViewModeler
}

func (m *RequestMiddleware) Handle(next http.Handler) http.Handler {
	logger := m.RootProvider.LoggerFactory.New("request")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(map[string]interface{}{
			"request.host":                    r.Host,
			"request.url":                     r.URL.String(),
			"request.header.host":             r.Header.Get("Host"),
			"request.header.x-forwarded-host": r.Header.Get("X-Forwarded-Host"),
			"request.header.x-original-host":  r.Header.Get("X-Original-Host"),
		}).Debug("serving request")
		appCtx, err := m.ConfigSource.ProvideContext(r.Context(), r)
		if err != nil {
			if errors.Is(err, configsource.ErrAppNotFound) {
				data := map[string]interface{}{
					"HTTPHost": string(m.HTTPHost),
				}
				baseViewModel := m.BaseViewModeler.ViewModel(r, w)
				viewmodels.Embed(data, baseViewModel)
				m.TemplateEngine.RenderStatus(w, r, http.StatusNotFound, TemplateWebAppNotFoundHTML, data)
			} else {
				logger.WithError(err).Error("failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		ap := m.RootProvider.NewAppProvider(r.Context(), appCtx)
		r = r.WithContext(withProvider(r.Context(), ap))

		{
			key := otelauthgear.AttributeKeyProjectID
			val := key.String(string(appCtx.Config.AppConfig.ID))
			r = r.WithContext(context.WithValue(r.Context(), key, val))
		}

		next.ServeHTTP(w, r)
	})
}
