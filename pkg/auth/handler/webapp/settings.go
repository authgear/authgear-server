package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeAuthUISettingsHTML string = "auth_ui_settings.html"
)

var TemplateAuthUISettingsHTML = template.T{
	Type:                    TemplateItemTypeAuthUISettingsHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
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
