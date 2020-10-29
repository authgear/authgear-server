package model

import (
	"time"
)

type CollaboratorRole string

const (
	CollaboratorRoleOwner  CollaboratorRole = "owner"
	CollaboratorRoleEditor CollaboratorRole = "editor"
)

type Collaborator struct {
	ID        string           `json:"id"`
	AppID     string           `json:"appID"`
	UserID    string           `json:"userID"`
	CreatedAt time.Time        `json:"createdAt"`
	Role      CollaboratorRole `json:"role"`
}

type CollaboratorInvitation struct {
	ID           string    `json:"id"`
	AppID        string    `json:"appID"`
	InvitedBy    string    `json:"invitedBy"`
	InviteeEmail string    `json:"inviteeEmail"`
	Code         string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpireAt     time.Time `json:"expireAt"`
}
