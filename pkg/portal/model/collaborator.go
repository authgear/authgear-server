package model

import (
	"time"
)

type CollaboratorRole string

const (
	CollaboratorRoleOwner  CollaboratorRole = "owner"
	CollaboratorRoleEditor CollaboratorRole = "editor"
)

// collaboratorRoleLevels indicates the general access level of roles. Lower
// level means more privileges.
var collaboratorRoleLevels = map[CollaboratorRole]int{
	CollaboratorRoleOwner:  1,
	CollaboratorRoleEditor: 2,
}

func (r CollaboratorRole) Level() int {
	level, ok := collaboratorRoleLevels[r]
	if !ok {
		panic("collaborator: unknown role " + string(r))
	}
	return level
}

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
