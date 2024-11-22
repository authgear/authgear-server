package basic

import (
	"net/http"
)

func UseRequestContext() {
	r := &http.Request{}
	_ = r.Context() // want `Unvetted usage of request.Context is forbidden.`
}
