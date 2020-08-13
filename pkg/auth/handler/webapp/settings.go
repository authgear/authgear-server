package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

const (
	TemplateItemTypeAuthUISettingsHTML config.TemplateItemType = "auth_ui_settings.html"
)

var TemplateAuthUISettingsHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISettingsHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

func ConfigureSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/settings")
}

type SettingsHandler struct {
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		baseViewModel := h.BaseViewModel.ViewModel(r, nil)

		data := map[string]interface{}{}

		viewmodels.Embed(data, baseViewModel)

		h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISettingsHTML, data)
		return
	}
}
