package dependency

import (
	"encoding/json"
	"time"

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
	now := time.Now().UTC()
	var userProfileBytes []byte
	userProfileBytes, err = json.Marshal(userProfile)
	if err != nil {
		return
	}

	builder := u.sqlBuilder.Insert(u.sqlBuilder.FullTableName("user_profile")).Columns(
		"user_id",
		"created_at",
		"created_by",
		"updated_at",
		"updated_by",
		"data",
	).Values(
		userID,
		now,
		userID,
		now,
		userID,
		userProfileBytes,
	)

	_, err = u.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	return
}

func (u UserProfileStoreImpl) GetUserProfile(userID string, userProfile *map[string]interface{}) (err error) {
	return
}
