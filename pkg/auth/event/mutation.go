package event

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Mutations struct {
	IsDisabled         *bool             `json:"is_disabled,omitempty"`
	VerifyInfo         *map[string]bool  `json:"verify_info,omitempty"`
	IsManuallyVerified *bool             `json:"is_manually_verified,omitempty"`
	IsComputedVerified *bool             `json:"-"`
	Metadata           *userprofile.Data `json:"metadata,omitempty"`
}

func (mutations Mutations) IsNoop() bool {
	return mutations == Mutations{}
}

func (mutations Mutations) WithMutationsApplied(newMutations Mutations) Mutations {
	if newMutations.IsDisabled != nil {
		mutations.IsDisabled = newMutations.IsDisabled
	}
	if newMutations.VerifyInfo != nil {
		mutations.VerifyInfo = newMutations.VerifyInfo
	}
	if newMutations.IsManuallyVerified != nil {
		mutations.IsManuallyVerified = newMutations.IsManuallyVerified
	}
	if newMutations.IsComputedVerified != nil {
		mutations.IsComputedVerified = newMutations.IsComputedVerified
	}
	if newMutations.Metadata != nil {
		mutations.Metadata = newMutations.Metadata
	}
	return mutations
}

func (mutations Mutations) ApplyToUser(user *model.User) {
	if mutations.IsDisabled != nil {
		user.Disabled = *mutations.IsDisabled
	}
	if mutations.VerifyInfo != nil {
		user.VerifyInfo = *mutations.VerifyInfo
	}
	if mutations.IsManuallyVerified != nil {
		user.ManuallyVerified = *mutations.IsManuallyVerified
	}
	if mutations.IsComputedVerified != nil {
		user.Verified = *mutations.IsComputedVerified || user.ManuallyVerified
	}
	if mutations.Metadata != nil {
		user.Metadata = *mutations.Metadata
	}
}
