package userverify

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Store interface {
	CreateVerifyCode(code *VerifyCode) error
	UpdateVerifyCode(code *VerifyCode) error
	GetVerifyCodeByCode(code string, vCode *VerifyCode) error
}

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry

	expiry int64
}

func newStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *storeImpl {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (s *storeImpl) CreateVerifyCode(code *VerifyCode) (err error) {
	builder := s.sqlBuilder.Insert(s.sqlBuilder.FullTableName("verify_code")).Columns(
		"id",
		"user_id",
		"record_key",
		"record_value",
		"code",
		"consumed",
		"created_at",
	).Values(
		code.ID,
		code.UserID,
		code.RecordKey,
		code.RecordValue,
		code.Code,
		code.Consumed,
		code.CreatedAt,
	)

	s.updateCodeExpiry(code)

	_, err = s.sqlExecutor.ExecWith(builder)
	return
}

func (s *storeImpl) UpdateVerifyCode(code *VerifyCode) (err error) {
	builder := s.sqlBuilder.Update(s.sqlBuilder.FullTableName("verify_code")).
		Set("user_id", code.UserID).
		Set("record_key", code.RecordKey).
		Set("record_value", code.RecordValue).
		Set("code", code.Code).
		Set("consumed", code.Consumed).
		Set("created_at", code.CreatedAt).
		Where("id = ?", code.ID)

	if _, err = s.sqlExecutor.ExecWith(builder); err != nil {
		return
	}

	s.updateCodeExpiry(code)

	return
}

func (s *storeImpl) GetVerifyCodeByCode(code string, vCode *VerifyCode) (err error) {
	builder := s.sqlBuilder.Select(
		"id",
		"user_id",
		"record_key",
		"record_value",
		"consumed",
		"created_at",
	).
		From(s.sqlBuilder.FullTableName("verify_code")).
		Where("code = ?", code).
		OrderBy("created_at desc")
	scanner := s.sqlExecutor.QueryRowWith(builder)

	var id string
	var userID string
	var recordKey string
	var recordValue string
	var consumed bool
	var createdAt time.Time
	err = scanner.Scan(
		&id,
		&userID,
		&recordKey,
		&recordValue,
		&consumed,
		&createdAt,
	)

	if err != nil {
		return
	}

	vCode.ID = id
	vCode.UserID = userID
	vCode.RecordKey = recordKey
	vCode.RecordValue = recordValue
	vCode.Code = code
	vCode.Consumed = consumed
	vCode.CreatedAt = createdAt

	return
}

func (s *storeImpl) updateCodeExpiry(code *VerifyCode) {
	if s.expiry > 0 {
		expireAt := code.CreatedAt.Add(time.Second * time.Duration(s.expiry))
		code.expireAt = &expireAt
	} else {
		code.expireAt = nil
	}
}
