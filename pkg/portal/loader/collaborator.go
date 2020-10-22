package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type CollaboratorLoaderCollaboratorService interface {
	GetManyCollaborators(ids []string) ([]*model.Collaborator, error)
	GetManyInvitations(ids []string) ([]*model.CollaboratorInvitation, error)
}

type CollaboratorLoader struct {
	*graphqlutil.DataLoader `wire:"-"`
	CollaboratorService     CollaboratorLoaderCollaboratorService
	Authz                   AuthzService
}

func NewCollaboratorLoader(
	collaboratorService CollaboratorLoaderCollaboratorService,
	authz AuthzService,
) *CollaboratorLoader {
	l := &CollaboratorLoader{
		CollaboratorService: collaboratorService,
		Authz:               authz,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *CollaboratorLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	collaborators, err := l.CollaboratorService.GetManyCollaborators(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.Collaborator)
	for _, domain := range collaborators {
		entityMap[domain.ID] = domain
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		_, err := l.Authz.CheckAccessOfViewer(entity.AppID)
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}

type CollaboratorInvitationLoader struct {
	*graphqlutil.DataLoader `wire:"-"`
	CollaboratorService     CollaboratorLoaderCollaboratorService
	Authz                   AuthzService
}

func NewCollaboratorInvitationLoader(
	collaboratorService CollaboratorLoaderCollaboratorService,
	authz AuthzService,
) *CollaboratorInvitationLoader {
	l := &CollaboratorInvitationLoader{
		CollaboratorService: collaboratorService,
		Authz:               authz,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *CollaboratorInvitationLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	invitations, err := l.CollaboratorService.GetManyInvitations(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.CollaboratorInvitation)
	for _, domain := range invitations {
		entityMap[domain.ID] = domain
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		_, err := l.Authz.CheckAccessOfViewer(entity.AppID)
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}
