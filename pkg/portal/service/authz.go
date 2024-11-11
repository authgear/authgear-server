package service

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

var ErrForbidden = apierrors.Forbidden.WithReason("Forbidden").New("forbidden")
var ErrUnauthenticated = apierrors.Unauthorized.WithReason("Unauthenticated").New("unauthenticated")

type AuthzConfigService interface {
	GetStaticAppIDs() ([]string, error)
}

type AuthzCollaboratorService interface {
	NewCollaborator(appID string, userID string, role model.CollaboratorRole) *model.Collaborator

	CreateCollaborator(ctx context.Context, c *model.Collaborator) error
	ListCollaboratorsByUser(ctx context.Context, userID string) ([]*model.Collaborator, error)
	GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error)
}

type AuthzService struct {
	Configs       AuthzConfigService
	Collaborators AuthzCollaboratorService
}

// ListAuthorizedApps calls other services that acquires connection themselves.
func (s *AuthzService) ListAuthorizedApps(ctx context.Context, userID string) ([]string, error) {
	appIDs, err := s.Configs.GetStaticAppIDs()
	if errors.Is(err, ErrGetStaticAppIDsNotSupported) {
		var cs []*model.Collaborator
		cs, err = s.Collaborators.ListCollaboratorsByUser(ctx, userID)
		if err == nil {
			appIDs = make([]string, len(cs))
			for i, c := range cs {
				appIDs[i] = c.AppID
			}
		}
	}

	if err != nil {
		return nil, err

	}

	return appIDs, nil
}

// AddAuthorizedUser assume acquired connection.
func (s *AuthzService) AddAuthorizedUser(ctx context.Context, appID string, userID string, role model.CollaboratorRole) error {
	c := s.Collaborators.NewCollaborator(appID, userID, role)
	return s.Collaborators.CreateCollaborator(ctx, c)
}

// CheckAccessOfViewer calls other services that acquires connection themselves.
func (s *AuthzService) CheckAccessOfViewer(ctx context.Context, appID string) (userID string, err error) {
	sessionInfo := session.GetValidSessionInfo(ctx)
	if sessionInfo == nil {
		err = ErrUnauthenticated
		return
	}

	userID = sessionInfo.UserID
	_, err = s.Collaborators.GetCollaboratorByAppAndUser(ctx, appID, userID)
	if errors.Is(err, ErrCollaboratorNotFound) {
		err = ErrForbidden
		return
	} else if err != nil {
		return
	}

	return
}
