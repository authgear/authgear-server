package settingsaction

import "net/http"

const (
	QUERY_SETTINGS_ACTION_ID = "x_settings_action_id"
)

func GetSettingsActionID(r *http.Request) string {
	return r.URL.Query().Get(QUERY_SETTINGS_ACTION_ID)
}
