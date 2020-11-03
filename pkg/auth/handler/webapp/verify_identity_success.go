package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebVerifyIdentitySuccessHTML = template.RegisterHTML(
	"web/verify_identity_success.html",
	components...,
)

func ConfigureVerifyIdentitySuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/verify_identity/success")
}

type VerifyIdentitySuccessHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *VerifyIdentitySuccessHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *VerifyIdentitySuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			data, err := h.GetData(r, w)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebVerifyIdentitySuccessHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
