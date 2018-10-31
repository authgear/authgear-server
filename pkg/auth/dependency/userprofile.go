package dependency

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type UserProfileStore interface {
	CreateUserProfile(userID string, userProfile map[string]interface{}) error
	GetUserProfile(userID string, userProfile *map[string]interface{}) error
}

type UserProfileStoreImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func NewUserProfileStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *UserProfileStoreImpl {
	return &UserProfileStoreImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (u UserProfileStoreImpl) CreateUserProfile(userID string, userProfile map[string]interface{}) (err error) {
	return
}

func (u UserProfileStoreImpl) GetUserProfile(userID string, userProfile *map[string]interface{}) (err error) {
	return
}
