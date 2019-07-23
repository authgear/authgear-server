package event

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

const (
	BeforeUserUpdate Type = "before_user_update"
	AfterUserUpdate  Type = "after_user_update"
)

const UserUpdateEventVersion int32 = 1

type UserUpdateReason string

const (
	UserUpdateReasonUpdateMetadata = "update-metadata"
	UserUpdateReasonUpdateIdentity = "update-identity"
	UserUpdateReasonVerification   = "verification"
	UserUpdateReasonAdministrative = "administrative"
)

type UserUpdateEvent struct {
	Reason     UserUpdateReason  `json:"reason"`
	IsDisabled *bool             `json:"is_disabled,omitempty"`
	IsVerified *bool             `json:"is_verified,omitempty"`
	VerifyInfo *map[string]bool  `json:"verify_info,omitempty"`
	Metadata   *userprofile.Data `json:"metadata,omitempty"`
	User       *model.User       `json:"user"`
}

func (UserUpdateEvent) Version() int32 {
	return 1
}
