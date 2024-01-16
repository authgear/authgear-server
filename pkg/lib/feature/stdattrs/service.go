package stdattrs

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type UserQueries interface {
	GetRaw(userID string) (*user.User, error)
}

type UserStore interface {
	UpdateStandardAttributes(userID string, stdAttrs map[string]interface{}) error
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type Service struct {
	UserProfileConfig *config.UserProfileConfig
	ServiceNoEvent    *ServiceNoEvent
	Identities        IdentityService
	UserQueries       UserQueries
	UserStore         UserStore
	Events            EventService
}

func (s *Service) PopulateStandardAttributes(userID string, iden *identity.Info) error {
	user, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	stdAttrsFromIden := stdattrs.T(iden.AllStandardClaims()).NonIdentityAware()
	originalStdAttrs := stdattrs.T(user.StandardAttributes)
	stdAttrs := originalStdAttrs.MergedWith(stdAttrsFromIden)

	err = s.UserStore.UpdateStandardAttributes(userID, stdAttrs.ToClaims())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateStandardAttributesWithList(role accesscontrol.Role, userID string, attrs attrs.List) error {
	user, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	originalStdAttrs := stdattrs.T(user.StandardAttributes)
	stdAttrs, err := originalStdAttrs.MergedWithList(attrs)
	if err != nil {
		return err
	}

	err = s.ServiceNoEvent.UpdateStandardAttributes(role, userID, stdAttrs)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PopulateIdentityAwareStandardAttributes(userID string) (err error) {
	return s.ServiceNoEvent.PopulateIdentityAwareStandardAttributes(userID)
}

func (s *Service) UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error {
	err := s.ServiceNoEvent.UpdateStandardAttributes(role, userID, stdAttrs)
	if err != nil {
		return err
	}

	eventPayloads := []event.Payload{
		&blocking.UserProfilePreUpdateBlockingEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			},
			AdminAPI: role == accesscontrol.RoleGreatest,
		},
		&nonblocking.UserProfileUpdatedEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			},
			AdminAPI: role == accesscontrol.RoleGreatest,
		},
	}

	for _, eventPayload := range eventPayloads {
		err = s.Events.DispatchEventOnCommit(eventPayload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error) {
	return s.ServiceNoEvent.DeriveStandardAttributes(role, userID, updatedAt, attrs)
}
