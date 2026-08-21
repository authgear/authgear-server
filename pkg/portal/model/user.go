package model

type User struct {
	ID                string `json:"id"`
	Email             string `json:"email,omitempty"`
	FormattedName     string `json:"formattedName,omitempty"`
	Picture           string `json:"picture,omitempty"`
	ProjectQuota      *int   `json:"projectQuota,omitempty"`
	ProjectOwnerCount int    `json:"projectOwnerCount,"`
	GeoIPCountryCode  string `json:"geoIPCountryCode,omitempty"`
}
