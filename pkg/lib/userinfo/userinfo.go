package userinfo

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type UserInfo struct {
	User              *model.User `json:"user,omitempty"`
	EffectiveRoleKeys []string    `json:"effective_role_keys,omitempty"`
}

type RolesAndGroupsQueries interface {
	ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
}

type UserQueries interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type UserInfoService struct {
	UserQueries           UserQueries
	RolesAndGroupsQueries RolesAndGroupsQueries
}

func (s *UserInfoService) GetUserInfo(ctx context.Context, userID string, role accesscontrol.Role) (*UserInfo, error) {
	u, err := s.UserQueries.Get(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	roles, err := s.RolesAndGroupsQueries.ListEffectiveRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleKeys := make([]string, len(roles))
	for i := range roles {
		roleKeys[i] = roles[i].Key
	}

	return &UserInfo{
		User:              u,
		EffectiveRoleKeys: roleKeys,
	}, nil
}
