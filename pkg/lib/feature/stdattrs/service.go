package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
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
	Get(userID string, role accesscontrol.Role) (*model.User, error)
}

type UserStore interface {
	UpdateStandardAttributes(userID string, stdAttrs map[string]interface{}) error
}

type EventService interface {
	DispatchEvent(payload event.Payload) error
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

	stdAttrsFromIden := stdattrs.T(iden.Claims).NonIdentityAware()
	originalStdAttrs := stdattrs.T(user.StandardAttributes)
	stdAttrs := originalStdAttrs.MergedWith(stdAttrsFromIden)

	err = s.UserStore.UpdateStandardAttributes(userID, stdAttrs.ToClaims())
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

	user, err := s.UserQueries.Get(userID, config.RolePortalUI)
	if err != nil {
		return err
	}

	eventPayloads := []event.Payload{
		&blocking.UserProfilePreUpdateBlockingEventPayload{
			User:     *user,
			AdminAPI: role == config.RolePortalUI,
		},
		&nonblocking.UserProfileUpdatedEventPayload{
			User:     *user,
			AdminAPI: role == config.RolePortalUI,
		},
	}

	for _, eventPayload := range eventPayloads {
		err = s.Events.DispatchEvent(eventPayload)
		if err != nil {
			return err
		}
	}

	return nil
}
