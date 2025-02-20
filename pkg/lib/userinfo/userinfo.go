package userinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

var errCacheMiss = errors.New("cache miss")

var ttl = duration.PerMinute

type UserInfo struct {
	User              *model.User `json:"user,omitempty"`
	EffectiveRoleKeys []string    `json:"effective_role_keys"`
}

type RolesAndGroupsQueries interface {
	ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
}

type UserQueries interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type UserInfoService struct {
	Redis                 *appredis.Handle
	AppID                 config.AppID
	UserQueries           UserQueries
	RolesAndGroupsQueries RolesAndGroupsQueries
}

func (s *UserInfoService) GetUserInfoGreatest(ctx context.Context, userID string) (*UserInfo, error) {
	return s.getUserInfo(ctx, userID, accesscontrol.RoleGreatest)
}

func (s *UserInfoService) GetUserInfoBearer(ctx context.Context, userID string) (*UserInfo, error) {
	return s.getUserInfo(ctx, userID, config.RoleBearer)
}

func (s *UserInfoService) getUserInfo(ctx context.Context, userID string, role accesscontrol.Role) (*UserInfo, error) {
	cached, err := s.getUserInfoFromCache(ctx, userID, role)
	if err != nil {
		if !errors.Is(err, errCacheMiss) {
			return nil, err
		}

		fresh, err := s.getUserInfoFromDatabase(ctx, userID, role)
		if err != nil {
			return nil, err
		}

		err = s.cacheUserInfo(ctx, userID, role, fresh)
		if err != nil {
			return nil, err
		}

		return fresh, nil
	}

	return cached, nil
}

func (s *UserInfoService) getUserInfoFromDatabase(ctx context.Context, userID string, role accesscontrol.Role) (*UserInfo, error) {
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

func (s *UserInfoService) getUserInfoFromCache(ctx context.Context, userID string, role accesscontrol.Role) (*UserInfo, error) {
	key := cacheKey(s.AppID, userID, role)

	var userInfo UserInfo
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return errCacheMiss
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &userInfo)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *UserInfoService) cacheUserInfo(ctx context.Context, userID string, role accesscontrol.Role, fresh *UserInfo) error {
	key := cacheKey(s.AppID, userID, role)
	data, err := json.Marshal(fresh)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(ctx, key, data, ttl).Result()
		return err
	})
}

func cacheKey(appID config.AppID, userID string, role accesscontrol.Role) string {
	return fmt.Sprintf("app:%s:userinfo:%s:%s", appID, userID, role)
}
