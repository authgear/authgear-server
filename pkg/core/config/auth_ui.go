package config

//go:generate msgp -tests=false
type AuthUIConfiguration struct {
	CSS                string                                 `json:"css,omitempty" yaml:"css" msg:"css"`
	CountryCallingCode *AuthUICountryCallingCodeConfiguration `json:"country_calling_code,omitempty" yaml:"country_calling_code" msg:"country_calling_code" default_zero_value:"true"`
	Metadata           AuthUIMetadataConfiguration            `json:"metadata,omitempty" yaml:"metadata" msg:"metadata" default_zero_value:"true"`
}

type AuthUICountryCallingCodeConfiguration struct {
	Values  []string `json:"values,omitempty" yaml:"values" msg:"values"`
	Default string   `json:"default,omitempty" yaml:"default" msg:"default"`
}

type AuthUIMetadataConfiguration map[string]interface{}
