package pq

import (
	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type storeImpl struct {
	mfaConfig    config.MFAConfiguration
	sqlBuilder   db.SQLBuilder
	sqlExecutor  db.SQLExecutor
	timeProvider time.Provider
}

func NewStore(
	mfaConfig config.MFAConfiguration,
	sqlBuilder db.SQLBuilder,
	sqlExecutor db.SQLExecutor,
	timeProvider time.Provider,
) mfa.Store {
	return &storeImpl{
		mfaConfig:    mfaConfig,
		sqlBuilder:   sqlBuilder,
		sqlExecutor:  sqlExecutor,
		timeProvider: timeProvider,
	}
}

func (s *storeImpl) GetRecoveryCode(userID string) (output []mfa.RecoveryCodeAuthenticator, err error) {
	builder := s.sqlBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"a.type",
			"arc.code",
			"arc.created_at",
			"arc.consumed",
		).
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_recovery_code"),
			"arc",
			"a.id = arc.id",
		).
		Where("a.user_id = ?", userID)
	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var a mfa.RecoveryCodeAuthenticator
		err = rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Type,
			&a.Code,
			&a.CreatedAt,
			&a.Consumed,
		)
		if err != nil {
			return
		}
		output = append(output, a)
	}

	return
}

func (s *storeImpl) GenerateRecoveryCode(userID string) ([]mfa.RecoveryCodeAuthenticator, error) {
	old, err := s.GetRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, a := range old {
		ids = append(ids, a.ID)
	}

	if len(ids) > 0 {
		q1 := s.sqlBuilder.Tenant().
			Delete(s.sqlBuilder.FullTableName("authenticator_recovery_code")).
			Where("id = ANY (?)", pq.Array(ids))

		_, err = s.sqlExecutor.ExecWith(q1)
		if err != nil {
			return nil, err
		}

		q2 := s.sqlBuilder.Tenant().
			Delete(s.sqlBuilder.FullTableName("authenticator")).
			Where("id = ANY (?)", pq.Array(ids))

		_, err = s.sqlExecutor.ExecWith(q2)
		if err != nil {
			return nil, err
		}
	}

	now := s.timeProvider.NowUTC()
	var output []mfa.RecoveryCodeAuthenticator
	for i := 0; i < s.mfaConfig.RecoveryCode.Count; i++ {
		a := mfa.RecoveryCodeAuthenticator{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      coreAuth.AuthenticatorTypeRecoveryCode,
			Code:      mfa.GenerateRandomRecoveryCode(),
			CreatedAt: now,
			Consumed:  false,
		}

		q3 := s.sqlBuilder.Tenant().
			Insert(s.sqlBuilder.FullTableName("authenticator")).
			Columns(
				"id",
				"type",
				"user_id",
			).
			Values(
				a.ID,
				a.Type,
				a.UserID,
			)
		_, err = s.sqlExecutor.ExecWith(q3)
		if err != nil {
			return nil, err
		}

		q4 := s.sqlBuilder.Tenant().
			Insert(s.sqlBuilder.FullTableName("authenticator_recovery_code")).
			Columns(
				"id",
				"code",
				"created_at",
				"consumed",
			).
			Values(
				a.ID,
				a.Code,
				a.CreatedAt,
				a.Consumed,
			)
		_, err := s.sqlExecutor.ExecWith(q4)
		if err != nil {
			return nil, err
		}

		output = append(output, a)
	}
	return output, nil
}

var (
	_ mfa.Store = &storeImpl{}
)
