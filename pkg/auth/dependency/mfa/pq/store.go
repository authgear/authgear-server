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

func (s *storeImpl) ListAuthenticators(userID string) ([]interface{}, error) {
	q1 := s.sqlBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"a.type",
			"at.activated",
			"at.created_at",
			"at.activated_at",
			"at.secret",
			"at.display_name",
		).
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_totp"),
			"at",
			"a.id = at.id",
		).
		Where("a.user_id = ? AND at.activated = TRUE", userID)
	rows1, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return nil, err
	}
	defer rows1.Close()

	var totps []mfa.TOTPAuthenticator
	for rows1.Next() {
		var a mfa.TOTPAuthenticator
		var activatedAt pq.NullTime
		err = rows1.Scan(
			&a.ID,
			&a.UserID,
			&a.Type,
			&a.Activated,
			&a.CreatedAt,
			&activatedAt,
			&a.Secret,
			&a.DisplayName,
		)
		if err != nil {
			return nil, err
		}
		if activatedAt.Valid {
			a.ActivatedAt = &activatedAt.Time
		}
		totps = append(totps, a)
	}

	q2 := s.sqlBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"a.type",
			"ao.activated",
			"ao.created_at",
			"ao.activated_at",
			"ao.channel",
			"ao.phone",
			"ao.email",
		).
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_oob"),
			"ao",
			"a.id = ao.id",
		).
		Where("a.user_id = ? AND ao.activated = TRUE", userID)
	rows2, err := s.sqlExecutor.QueryWith(q2)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	var oobs []mfa.OOBAuthenticator
	for rows2.Next() {
		var a mfa.OOBAuthenticator
		var activatedAt pq.NullTime
		err = rows2.Scan(
			&a.ID,
			&a.UserID,
			&a.Type,
			&a.Activated,
			&a.CreatedAt,
			&activatedAt,
			&a.Channel,
			&a.Phone,
			&a.Email,
		)
		if err != nil {
			return nil, err
		}
		if activatedAt.Valid {
			a.ActivatedAt = &activatedAt.Time
		}
		oobs = append(oobs, a)
	}

	output := []interface{}{}
	for _, a := range totps {
		output = append(output, a)
	}
	for _, a := range oobs {
		output = append(output, a)
	}
	return output, nil
}

func (s *storeImpl) CreateTOTP(a *mfa.TOTPAuthenticator) error {
	q1 := s.sqlBuilder.Tenant().
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
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}

	q2 := s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("authenticator_totp")).
		Columns(
			"id",
			"activated",
			"created_at",
			"activated_at",
			"secret",
			"display_name",
		).
		Values(
			a.ID,
			a.Activated,
			a.CreatedAt,
			a.ActivatedAt,
			a.Secret,
			a.DisplayName,
		)
	_, err = s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return err
	}

	return nil
}

var (
	_ mfa.Store = &storeImpl{}
)
