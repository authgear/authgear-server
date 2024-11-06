package webapp

import (
	"encoding/json"
	"net/http"

	"strings"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateTurboErrorHTML = template.RegisterHTML(
	"web/turbo_error.html",
	Components...,
)

type ResponseWriter struct {
	JSONResponseWriter JSONResponseWriter
	Renderer           Renderer
}

func (w *ResponseWriter) WriteResponse(rw http.ResponseWriter, req *http.Request, resp *api.Response) {
	const turboStreamMedia = "text/vnd.turbo-stream.html"
	accept := req.Header.Get("Accept")
	if strings.Contains(accept, turboStreamMedia) && resp.Error != nil {
		data := w.PrepareData(resp.Error)
		w.Renderer.RenderStatus(rw, req, http.StatusInternalServerError, TemplateTurboErrorHTML, data)
		return
	}

	w.JSONResponseWriter.WriteResponse(rw, resp)
}

func (w *ResponseWriter) PrepareData(err error) map[string]interface{} {
	apiError := apierrors.AsAPIError(err)
	b, err := json.Marshal(struct {
		Error *apierrors.APIError `json:"error"`
	}{apiError})
	if err != nil {
		panic(err)
	}
	var eJSON map[string]interface{}
	err = json.Unmarshal(b, &eJSON)
	if err != nil {
		panic(err)
	}
	return map[string]interface{}{"Error": eJSON["error"]}
}
