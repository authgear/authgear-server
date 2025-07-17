package model

type Resource struct {
	Meta
	URI  string  `json:"uri"`
	Name *string `json:"name,omitzero"`
}

type Scope struct {
	Meta
	ResourceID  string  `json:"resource_id"`
	Scope       string  `json:"scope"`
	Description *string `json:"description,omitzero"`
}
