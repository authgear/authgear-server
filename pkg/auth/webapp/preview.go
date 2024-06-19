package webapp

import "net/http"

const PreviewQueryKey = "x_preview"

type PreviewMode string

const (
	PreviewModeInline = "inline"
)

func IsPreviewModeInline(r *http.Request) bool {
	return r.Method == http.MethodGet && r.URL.Query().Get(PreviewQueryKey) == PreviewModeInline
}
