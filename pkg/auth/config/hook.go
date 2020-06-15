package config

type HookConfig struct {
	SyncTimeout      DurationSeconds     `json:"sync_hook_timeout_seconds,omitempty"`
	SyncTotalTimeout DurationSeconds     `json:"sync_hook_total_timeout_seconds,omitempty"`
	Handlers         []HookHandlerConfig `json:"handlers,omitempty"`
}

type HookHandlerConfig struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}
