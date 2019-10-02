package userprofile

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func newUserProfileStore(builder db.SQLBuilder, executor db.SQLExecutor, loggerFactory logging.Factory) *storeImpl {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      loggerFactory.NewLogger("user-profile"),
	}
}

func (u storeImpl) CreateUserProfile(userID string, data Data) (profile UserProfile, err error) {
	now := timeNow()
	var dataBytes []byte
	dataBytes, err = json.Marshal(data)
	if err != nil {
		return
	}

	builder := u.sqlBuilder.Tenant().
		Insert(u.sqlBuilder.FullTableName("user_profile")).
		Columns(
			"user_id",
			"created_at",
			"created_by",
			"updated_at",
			"updated_by",
			"data",
		).
		Values(
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

	profile = u.toUserProfile(userID, data, now, now)
	return
}

func (u storeImpl) GetUserProfile(userID string) (profile UserProfile, err error) {
	builder := u.sqlBuilder.Tenant().
		Select("created_at", "updated_at", "data").
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

	profile = u.toUserProfile(userID, data, createdAt, updatedAt)
	return
}

func (u storeImpl) UpdateUserProfile(userID string, data Data) (profile UserProfile, err error) {
	profile, err = u.GetUserProfile(userID)
	if err != nil {
		return
	}

	profile.Data = data

	var dataBytes []byte
	dataBytes, err = json.Marshal(profile.Data)
	if err != nil {
		return
	}

	now := timeNow()
	builder := u.sqlBuilder.Tenant().
		Update(u.sqlBuilder.FullTableName("user_profile")).
		Set("updated_at", now).
		Set("data", dataBytes).
		Where("user_id = ?", userID)

	_, err = u.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	profile = u.toUserProfile(userID, profile.Data, profile.CreatedAt, now)
	return
}

func (u storeImpl) toUserProfile(userID string, data Data, createdAt time.Time, updatedAt time.Time) UserProfile {
	return UserProfile{
		ID:        userID,
		CreatedAt: createdAt,
		CreatedBy: userID,
		UpdatedAt: updatedAt,
		UpdatedBy: userID,
		Data:      data,
	}
}
