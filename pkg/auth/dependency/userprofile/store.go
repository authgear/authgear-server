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

type userProfile struct {
	userID    string
	createdAt time.Time
	updatedAt time.Time
	data      map[string]interface{}
}

func newUserProfileStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *storeImpl {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (u storeImpl) CreateUserProfile(userID string, data Data) (profile UserProfile, err error) {
	now := timeNow()
	var dataBytes []byte
	dataBytes, err = json.Marshal(data)
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
		dataBytes,
	)

	_, err = u.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	profile = userProfile{
		userID:    userID,
		createdAt: now,
		updatedAt: now,
		data:      data,
	}.unmarshal()

	return
}

func (u storeImpl) GetUserProfile(userID string) (profile UserProfile, err error) {
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

	var data map[string]interface{}
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return
	}

	profile = userProfile{
		userID:    userID,
		createdAt: createdAt,
		updatedAt: updatedAt,
		data:      data,
	}.unmarshal()

	return
}

func (u userProfile) unmarshal() UserProfile {
	profile := make(map[string]interface{})

	profile["_id"] = "user/" + u.userID
	profile["_type"] = "record"
	profile["_recordID"] = u.userID
	profile["_recordType"] = "user"
	profile["_access"] = nil
	profile["_ownerID"] = u.userID
	profile["_created_at"] = u.createdAt
	profile["_created_by"] = u.userID
	profile["_updated_at"] = u.updatedAt
	profile["_updated_by"] = u.userID
	for k, v := range u.data {
		profile[k] = v
	}

	return profile
}
