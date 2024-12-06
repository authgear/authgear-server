package webapp

import (
	"net/http"
)

type SettingsImplementationSwitcherHandler struct {
	SettingV1 http.Handler
	SettingV2 http.Handler
}

func (h *SettingsImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.SettingV2.ServeHTTP(w, r)
}
