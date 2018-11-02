package userprofile

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func newUserProfileStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *storeImpl {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (u storeImpl) CreateUserProfile(userID string, userProfile map[string]interface{}) (err error) {
	now := timeNow()
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

func (u storeImpl) GetUserProfile(userID string, userProfile *map[string]interface{}) (err error) {
	builder := u.sqlBuilder.Select("created_at", "updated_at", "data").
		From(u.sqlBuilder.FullTableName("user_profile")).
		Where("user_id = ?", userID)
	scanner := u.sqlExecutor.QueryRowWith(builder)
	var createdAt time.Time
	var updatedAt time.Time
	var dataBytes []byte
	err = scanner.Scan(
		&createdAt,
		&updatedAt,
		&dataBytes,
	)

	if err != nil {
		return
	}

	// generate default record attributes
	err = json.Unmarshal(dataBytes, &userProfile)
	(*userProfile)["_id"] = "user/" + userID
	(*userProfile)["_type"] = "record"
	(*userProfile)["_recordID"] = userID
	(*userProfile)["_recordType"] = "user"
	(*userProfile)["_access"] = nil
	(*userProfile)["_ownerID"] = userID
	(*userProfile)["_created_at"] = createdAt
	(*userProfile)["_created_by"] = userID
	(*userProfile)["_updated_at"] = updatedAt
	(*userProfile)["_updated_by"] = userID

	return
}
