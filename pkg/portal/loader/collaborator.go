package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type CollaboratorService interface {
	ListCollaborators(appID string) ([]*model.Collaborator, error)
	GetCollaborator(id string) (*model.Collaborator, error)
	DeleteCollaborator(c *model.Collaborator) error

	ListInvitations(appID string) ([]*model.CollaboratorInvitation, error)
	GetInvitation(id string) (*model.CollaboratorInvitation, error)
	DeleteInvitation(i *model.CollaboratorInvitation) error
	SendInvitation(appID string, inviteeEmail string) (*model.CollaboratorInvitation, error)
	AcceptInvitation(code string) (*model.Collaborator, error)
}

type CollaboratorLoader struct {
	Collaborators CollaboratorService
	Authz         AuthzService
}

func (l *CollaboratorLoader) ListCollaborators(appID string) *graphqlutil.Lazy {
	_, err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

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

		_, err = l.Authz.CheckAccessOfViewer(c.AppID)
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

func (l *CollaboratorLoader) ListInvitations(appID string) *graphqlutil.Lazy {
	_, err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Collaborators.ListInvitations(appID)
	})
}

func (l *CollaboratorLoader) DeleteInvitation(id string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		i, err := l.Collaborators.GetInvitation(id)
		if err != nil {
			return nil, err
		}

		_, err = l.Authz.CheckAccessOfViewer(i.AppID)
		if err != nil {
			return nil, err
		}

		err = l.Collaborators.DeleteInvitation(i)
		if err != nil {
			return nil, err
		}

		return i, nil
	})
}

func (l *CollaboratorLoader) SendInvitation(appID string, inviteeEmail string) *graphqlutil.Lazy {
	_, err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		i, err := l.Collaborators.SendInvitation(appID, inviteeEmail)
		if err != nil {
			return nil, err
		}
		return i, nil
	})
}

func (l *CollaboratorLoader) AcceptInvitation(code string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		c, err := l.Collaborators.AcceptInvitation(code)
		if err != nil {
			return nil, err
		}
		return c, nil
	})
}
