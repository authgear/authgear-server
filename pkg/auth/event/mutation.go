package event

import "github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

type Mutations struct {
	IsDisabled *bool             `json:"is_disabled,omitempty"`
	VerifyInfo *map[string]bool  `json:"verify_info,omitempty"`
	Metadata   *userprofile.Data `json:"metadata,omitempty"`
}
