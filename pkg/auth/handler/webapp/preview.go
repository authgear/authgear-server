package webapp

import (
	"net/http"
)

type PreviewHandler struct {
	previewHandleFunc func() error
}

func (h *PreviewHandler) Preview(previewHandleFunc func() error) {
	h.previewHandleFunc = previewHandleFunc
}

func (h *PreviewHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if err := h.previewHandleFunc(); err != nil {
		h.renderError(err, w, r)
		return
	}
}

func (h *PreviewHandler) renderError(
	err error,
	w http.ResponseWriter,
	r *http.Request,
) {
	// Panic first. Will decide how to handle later
	panic(err)
}
