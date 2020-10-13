package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type CollaboratorService interface {
	ListCollaborators(appID string) ([]*model.Collaborator, error)
}

type CollaboratorLoader struct {
	Collaborators CollaboratorService
}

func (l *CollaboratorLoader) ListCollaborators(appID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Collaborators.ListCollaborators(appID)
	})
}
