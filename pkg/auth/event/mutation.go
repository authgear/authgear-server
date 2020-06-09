package event

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Mutations struct {
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

func (mutations Mutations) IsNoop() bool {
	return mutations == Mutations{}
}

func (mutations Mutations) WithMutationsApplied(newMutations Mutations) Mutations {
	if newMutations.Metadata != nil {
		mutations.Metadata = newMutations.Metadata
	}
	return mutations
}

func (mutations Mutations) ApplyToUser(user *model.User) {
	if mutations.Metadata != nil {
		user.Metadata = *mutations.Metadata
	}
}
