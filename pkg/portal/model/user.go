package model

type User struct {
	ID                string `json:"id"`
	Email             string `json:"email,omitempty"`
	ProjectQuota      *int   `json:"projectQuota,omitempty"`
	ProjectOwnerCount int    `json:"projectOwnerCount,"`
}
