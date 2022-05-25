package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
)

type ResponseWriter struct {
	JSONResponseWriter JSONResponseWriter
}

func (w *ResponseWriter) WriteResponse(rw http.ResponseWriter, req *http.Request, resp *api.Response) {
	w.JSONResponseWriter.WriteResponse(rw, resp)
}
