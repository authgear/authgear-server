package webapp

import (
	"net/http"
	"strings"
)

const InlinePreviewPathPrefix = "/preview/"

func IsInlinePreviewPageRequest(r *http.Request) bool {
	return r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, InlinePreviewPathPrefix)
}
