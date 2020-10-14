package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type CollaboratorService interface {
	ListCollaborators(appID string) ([]*model.Collaborator, error)
	GetCollaborator(id string) (*model.Collaborator, error)
	DeleteCollaborator(c *model.Collaborator) error
}

type CollaboratorLoader struct {
	Collaborators CollaboratorService
}

func (l *CollaboratorLoader) ListCollaborators(appID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Collaborators.ListCollaborators(appID)
	})
}

func (l *CollaboratorLoader) DeleteCollaborator(id string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		c, err := l.Collaborators.GetCollaborator(id)
		if err != nil {
			return nil, err
		}

		err = l.Collaborators.DeleteCollaborator(c)
		if err != nil {
			return nil, err
		}

		return c, nil
	})
}
