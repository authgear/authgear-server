package config

type HTTPConfig struct {
	Hosts         []string `json:"hosts,omitempty"`
	AllowsOrigins []string `json:"allowed_origins,omitempty"`
}
