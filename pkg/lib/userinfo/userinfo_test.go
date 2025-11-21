package userinfo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	gomock "github.com/golang/mock/gomock"
	goredis "github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestUserInfoSerialization(t *testing.T) {
	Convey("UserInfo serialization", t, func() {
		u := &UserInfo{
			User: &model.User{
				StandardAttributes: map[string]interface{}{},
				CustomAttributes:   map[string]interface{}{},
				Web3:               map[string]interface{}{},
				Roles:              []string{},
				Groups:             []string{},
			},
			EffectiveRoleKeys: []string{},
			Authenticators:    []model.UserInfoAuthenticator{},
		}

		b, err := json.Marshal(u)
		So(err, ShouldBeNil)

		var uu UserInfo
		err = json.Unmarshal(b, &uu)
		So(err, ShouldBeNil)

		So(uu.EffectiveRoleKeys, ShouldNotBeNil)
		So(uu.EffectiveRoleKeys, ShouldHaveLength, 0)

		So(uu.User.Roles, ShouldNotBeNil)
		So(uu.User.Roles, ShouldHaveLength, 0)

		So(uu.User.Groups, ShouldNotBeNil)
		So(uu.User.Groups, ShouldHaveLength, 0)

		So(uu.User.StandardAttributes, ShouldNotBeNil)
		So(uu.User.StandardAttributes, ShouldHaveLength, 0)

		So(uu.User.CustomAttributes, ShouldNotBeNil)
		So(uu.User.CustomAttributes, ShouldHaveLength, 0)

		So(uu.Authenticators, ShouldNotBeNil)
		So(uu.Authenticators, ShouldHaveLength, 0)
	})
}

func TestGetUserInfoBearer(t *testing.T) {
	Convey("GetUserInfoBearer", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mr := miniredis.RunT(t)

		client := goredis.NewClient(&goredis.Options{
			Addr: mr.Addr(),
		})
		defer client.Close()

		pool := redis.NewPool()
		So(pool, ShouldNotBeNil)

		redisConfig := &config.RedisEnvironmentConfig{}
		redisCredentials := &config.RedisCredentials{
			RedisURL: "redis://" + mr.Addr(),
		}

		hub := redis.NewHub(context.Background(), pool)
		rh := appredis.NewHandle(pool, hub, redisConfig, redisCredentials)

		userQueries := NewMockUserQueries(ctrl)
		rolesAndGroupsQueries := NewMockRolesAndGroupsQueries(ctrl)
		authenticatorService := NewMockUserInfoAuthenticatorService(ctrl)
		mfaService := NewMockUserInfoMFAService(ctrl)

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := now.Add(-1 * time.Hour)
		updatedAt := now.Add(-30 * time.Minute)

		user := &model.User{
			Meta: model.Meta{
				ID: "user-id",
			},
		}
		roles := []*model.Role{
			{
				Key: "role-1",
			},
		}
		authns := []*authenticator.Info{
			{
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Type:      model.AuthenticatorTypePassword,
				Kind:      model.AuthenticatorKindPrimary,
			},
			{
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Type:      model.AuthenticatorTypeOOBSMS,
				Kind:      model.AuthenticatorKindSecondary,
				OOBOTP:    &authenticator.OOBOTP{},
			},
			{
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Type:      model.AuthenticatorTypeOOBEmail,
				Kind:      model.AuthenticatorKindSecondary,
				OOBOTP:    &authenticator.OOBOTP{},
			},
			{
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Type:      model.AuthenticatorTypeTOTP,
				Kind:      model.AuthenticatorKindSecondary,
			},
		}

		userQueries.EXPECT().Get(gomock.Any(), "user-id", config.RoleBearer).Return(user, nil)
		rolesAndGroupsQueries.EXPECT().ListEffectiveRolesByUserID(gomock.Any(), "user-id").Return(roles, nil)
		authenticatorService.EXPECT().List(gomock.Any(), "user-id", gomock.Any()).Return(authns, nil)
		mfaService.EXPECT().ListRecoveryCodes(gomock.Any(), "user-id").Return([]*mfa.RecoveryCode{
			{
				Code: "some-code",
			},
		}, nil)

		s := &UserInfoService{
			Redis:                 rh,
			Clock:                 clock.NewMockClockAtTime(now),
			AppID:                 "test-app-id",
			UserQueries:           userQueries,
			RolesAndGroupsQueries: rolesAndGroupsQueries,
			AuthenticatorService:  authenticatorService,
			MFAService:            mfaService,
			AuthenticationConfig: &config.AuthenticationConfig{
				PrimaryAuthenticators: &[]model.AuthenticatorType{
					model.AuthenticatorTypePassword,
				},
				SecondaryAuthenticators: &[]model.AuthenticatorType{
					model.AuthenticatorTypeOOBEmail,
					model.AuthenticatorTypeOOBSMS,
					model.AuthenticatorTypeTOTP,
				},
			},
		}

		userInfo, err := s.GetUserInfoBearer(context.Background(), "user-id")
		So(err, ShouldBeNil)
		So(userInfo, ShouldResemble, &UserInfo{
			User:              user,
			EffectiveRoleKeys: []string{"role-1"},
			Authenticators: []model.UserInfoAuthenticator{
				{
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
					Type:      model.AuthenticatorTypePassword,
					Kind:      model.AuthenticatorKindPrimary,
				},
				{
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
					Type:      model.AuthenticatorTypeOOBSMS,
					Kind:      model.AuthenticatorKindSecondary,
				},
				{
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
					Type:      model.AuthenticatorTypeOOBEmail,
					Kind:      model.AuthenticatorKindSecondary,
				},
				{
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
					Type:      model.AuthenticatorTypeTOTP,
					Kind:      model.AuthenticatorKindSecondary,
				},
			},
			RecoveryCodeEnabled: true,
		})
	})
}
