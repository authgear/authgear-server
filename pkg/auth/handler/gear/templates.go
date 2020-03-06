package gear

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func AttachTemplatesHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/.skygear/templates.json").
		Handler(server.FactoryToHandler(&TemplatesHandlerFactory{})).
		Methods("GET")
}

type TemplatesHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f TemplatesHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &TemplatesHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

type TemplatesHandler struct {
	TemplateEngine *template.Engine `dependency:"TemplateEngine"`
}

func (h TemplatesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json")

	specs := make([]template.Spec, len(h.TemplateEngine.TemplateSpecs))
	i := 0
	for _, spec := range h.TemplateEngine.TemplateSpecs {
		specs[i] = spec
		i++
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Type < specs[j].Type })
	encoder.Encode(specs)
}
