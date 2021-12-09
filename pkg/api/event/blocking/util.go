package blocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

func ApplyMutations(user model.User, mutations event.Mutations) (out model.User, mutated bool) {
	if mutations.User.StandardAttributes != nil {
		user.StandardAttributes = mutations.User.StandardAttributes
		mutated = true
	}
	if mutations.User.CustomAttributes != nil {
		user.CustomAttributes = mutations.User.CustomAttributes
		mutated = true
	}

	out = user
	return
}

func GenerateFullMutations(user model.User) event.Mutations {
	return event.Mutations{
		User: event.UserMutations{
			StandardAttributes: user.StandardAttributes,
			CustomAttributes:   user.CustomAttributes,
		},
	}
}
