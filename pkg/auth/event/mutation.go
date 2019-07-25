package event

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Mutations struct {
	IsDisabled *bool             `json:"is_disabled,omitempty"`
	IsVerified *bool             `json:"is_verified,omitempty"`
	VerifyInfo *map[string]bool  `json:"verify_info,omitempty"`
	Metadata   *userprofile.Data `json:"metadata,omitempty"`
}

func (mutations Mutations) ApplyingMutations(newMutations Mutations) Mutations {
	if newMutations.IsDisabled != nil {
		mutations.IsDisabled = newMutations.IsDisabled
	}
	if newMutations.VerifyInfo != nil {
		mutations.VerifyInfo = newMutations.VerifyInfo
	}
	if newMutations.IsVerified != nil {
		mutations.IsVerified = newMutations.IsVerified
	}
	if newMutations.Metadata != nil {
		mutations.Metadata = newMutations.Metadata
	}
	return mutations
}

func (mutations Mutations) ApplyingToUser(user *model.User) *model.User {
	if user == nil {
		return nil
	}
	newUser := *user
	if mutations.IsDisabled != nil {
		newUser.Disabled = *mutations.IsDisabled
	}
	if mutations.VerifyInfo != nil {
		newUser.VerifyInfo = *mutations.VerifyInfo
	}
	if mutations.IsVerified != nil {
		newUser.Verified = *mutations.IsVerified
	}
	if mutations.Metadata != nil {
		newUser.Metadata = *mutations.Metadata
	}
	return &newUser
}
