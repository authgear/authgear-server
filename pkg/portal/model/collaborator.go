package model

import (
	"time"
)

type Collaborator struct {
	ID        string    `json:"id"`
	AppID     string    `json:"appID"`
	UserID    string    `json:"userID"`
	CreatedAt time.Time `json:"createdAt"`
}
