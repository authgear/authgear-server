package userinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=userinfo.go -destination=userinfo_mock_test.go -package userinfo

var errCacheMiss = errors.New("cache miss")

var ttl = duration.Short

var UserInfoCacheLogger = slogutil.NewLogger("userinfo-cache")

type UserInfo struct {
	User                    *model.User                   `json:"user,omitempty"`
	AccountAccountStaleFrom *time.Time                    `json:"account_status_stale_from,omitempty"`
	EffectiveRoleKeys       []string                      `json:"effective_role_keys"`
	Authenticators          []model.UserInfoAuthenticator `json:"authenticators"`
	RecoveryCodeEnabled     bool                          `json:"recovery_code_enabled"`
}

type RolesAndGroupsQueries interface {
	ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
}

type UserInfoAuthenticatorService interface {
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type UserInfoMFAService interface {
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
}

type UserQueries interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type UserInfoService struct {
	Redis                 *appredis.Handle
	Clock                 clock.Clock
	AppID                 config.AppID
	AuthenticationConfig  *config.AuthenticationConfig
	UserQueries           UserQueries
	RolesAndGroupsQueries RolesAndGroupsQueries
	AuthenticatorService  UserInfoAuthenticatorService
	MFAService            UserInfoMFAService
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

	authns, err := s.AuthenticatorService.List(ctx, userID, authenticator.FilterFunc(func(info *authenticator.Info) bool {
		switch info.Kind {
		case authenticator.KindPrimary:
			return slices.Contains(*s.AuthenticationConfig.PrimaryAuthenticators, info.Type)
		case authenticator.KindSecondary:
			return slices.Contains(*s.AuthenticationConfig.SecondaryAuthenticators, info.Type)
		default:
			return false
		}
	}))
	if err != nil {
		return nil, err
	}

	userinfoAuthens := []model.UserInfoAuthenticator{}
	for _, authn := range authns {
		userinfoAuthen := model.UserInfoAuthenticator{
			CreatedAt: authn.CreatedAt,
			UpdatedAt: authn.UpdatedAt,
			Kind:      authn.Kind,
			Type:      authn.Type,
		}

		userinfoAuthens = append(userinfoAuthens, userinfoAuthen)
	}

	recoveryCodes, err := s.MFAService.ListRecoveryCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		User:                    u,
		AccountAccountStaleFrom: u.AccountStatusStaleFrom,
		EffectiveRoleKeys:       roleKeys,
		Authenticators:          userinfoAuthens,
		RecoveryCodeEnabled:     len(recoveryCodes) > 0,
	}, nil
}

func (s *UserInfoService) getUserInfoFromCache(ctx context.Context, userID string, role accesscontrol.Role) (*UserInfo, error) {
	logger := UserInfoCacheLogger.GetLogger(ctx)

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
		if errors.Is(err, errCacheMiss) {
			logger.Debug(ctx, "userinfo cache miss")
		}
		return nil, err
	}

	// If account_status_stale_from is non-nil,
	// then it specifies the timestamp when the user info is considered as stale.
	if userInfo.AccountAccountStaleFrom != nil {
		now := s.Clock.NowUTC()
		if !now.Before(*userInfo.AccountAccountStaleFrom) {
			logger.Debug(ctx, "userinfo cache miss")
			return nil, errCacheMiss
		}
	}

	logger.Debug(ctx, "userinfo cache hit")
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

func (s *UserInfoService) PurgeUserInfo(ctx context.Context, userID string) error {
	keys := []string{
		cacheKey(s.AppID, userID, accesscontrol.RoleGreatest),
		cacheKey(s.AppID, userID, config.RoleBearer),
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, keys...).Result()
		return err
	})
}

func cacheKey(appID config.AppID, userID string, role accesscontrol.Role) string {
	return fmt.Sprintf("app:%s:userinfo:%s:%s", appID, userID, role)
}
