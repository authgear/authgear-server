package userverify

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Store interface {
	CreateVerifyCode(code *VerifyCode) error
	MarkConsumed(codeID string) error
	GetVerifyCodeByUser(userID string) (VerifyCode, error)
}

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
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
		"login_id_key",
		"login_id",
		"code",
		"consumed",
		"created_at",
	).Values(
		code.ID,
		code.UserID,
		code.LoginIDKey,
		code.LoginID,
		code.Code,
		code.Consumed,
		code.CreatedAt,
	)

	_, err = s.sqlExecutor.ExecWith(builder)
	return
}

func (s *storeImpl) MarkConsumed(codeID string) (err error) {
	builder := s.sqlBuilder.Update(s.sqlBuilder.FullTableName("verify_code")).
		Set("consumed", true).
		Where("id = ?", codeID)

	if _, err = s.sqlExecutor.ExecWith(builder); err != nil {
		return err
	}

	return
}

func (s *storeImpl) GetVerifyCodeByUser(userID string) (VerifyCode, error) {
	builder := s.sqlBuilder.Select(
		"id",
		"code",
		"user_id",
		"login_id_key",
		"login_id",
		"consumed",
		"created_at",
	).
		From(s.sqlBuilder.FullTableName("verify_code")).
		Where("user_id = ?", userID).
		OrderBy("created_at desc")
	scanner := s.sqlExecutor.QueryRowWith(builder)

	verifyCode := VerifyCode{}
	err := scanner.Scan(
		&verifyCode.ID,
		&verifyCode.Code,
		&verifyCode.UserID,
		&verifyCode.LoginIDKey,
		&verifyCode.LoginID,
		&verifyCode.Consumed,
		&verifyCode.CreatedAt,
	)

	return verifyCode, err
}

var _ Store = &storeImpl{}
