package loader

import "github.com/authgear/authgear-server/pkg/lib/interaction"

type InteractionService interface {
	Perform(intent interaction.Intent, input interface{}) (*interaction.Graph, error)
}
