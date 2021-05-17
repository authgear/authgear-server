package model

import (
	"time"
)

type SessionGroupType string

const (
	SessionGroupTypeUngrouped SessionGroupType = "ungrouped"
	SessionGroupTypeGrouped   SessionGroupType = "grouped"
)

type SessionGroup struct {
	Type           SessionGroupType
	DisplayName    string
	LastAccessedAt time.Time
	// OfflineGrantIDs is the list of offline grant IDs in this group.
	OfflineGrantIDs []string
	// Sessions is the list of sessions in this group.
	Sessions []*Session
}
