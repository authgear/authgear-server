package userprofile

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type MockUserProfileStoreImpl struct {
	Data        map[string]map[string]interface{}
	TimeNowfunc MockTimeNowfunc
}

type MockTimeNowfunc func() time.Time

func NewMockUserProfileStore() *MockUserProfileStoreImpl {
	return &MockUserProfileStoreImpl{
		Data:        map[string]map[string]interface{}{},
		TimeNowfunc: func() time.Time { return time.Time{} },
	}
}

func NewMockUserProfileStoreByData(data map[string]map[string]interface{}) *MockUserProfileStoreImpl {
	return &MockUserProfileStoreImpl{
		Data:        data,
		TimeNowfunc: func() time.Time { return time.Time{} },
	}
}

func (u MockUserProfileStoreImpl) CreateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (profile UserProfile, err error) {
	u.Data[userID] = data
	now := u.TimeNowfunc()
	profile = UserProfile{
		ID:        userID,
		CreatedAt: now,
		CreatedBy: userID,
		UpdatedAt: now,
		UpdatedBy: userID,
		Data:      data,
	}
	return
}

func (u MockUserProfileStoreImpl) GetUserProfile(userID string) (profile UserProfile, err error) {
	data := u.Data[userID]
	now := u.TimeNowfunc()
	profile = UserProfile{
		ID:        userID,
		CreatedAt: now,
		CreatedBy: userID,
		UpdatedAt: now,
		UpdatedBy: userID,
		Data:      data,
	}
	return
}

func (u MockUserProfileStoreImpl) UpdateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (profile UserProfile, err error) {
	u.Data[userID] = data
	now := u.TimeNowfunc()
	profile = UserProfile{
		ID:        userID,
		CreatedAt: now,
		CreatedBy: userID,
		UpdatedAt: now,
		UpdatedBy: userID,
		Data:      u.Data[userID],
	}
	return
}
