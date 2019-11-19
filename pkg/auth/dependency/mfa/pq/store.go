package pq

import (
	"database/sql"
	"sort"
	gotime "time"

	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type storeImpl struct {
	mfaConfig    *config.MFAConfiguration
	sqlBuilder   db.SQLBuilder
	sqlExecutor  db.SQLExecutor
	timeProvider time.Provider
}

func NewStore(
	mfaConfig *config.MFAConfiguration,
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

func sortRecoveryCodeAuthenticatorSlice(s []mfa.RecoveryCodeAuthenticator) {
	sort.Slice(s, func(i, j int) bool {
		a := s[i]
		b := s[j]
		return a.Code < b.Code
	})
}

func sortAuthenticatorSlice(s []mfa.Authenticator) {
	sort.Slice(s, func(i, j int) bool {
		a := s[i]
		b := s[j]
		return a.GetActivatedAt().After(*b.GetActivatedAt())
	})
}

func (s *storeImpl) scanTOTPAuthenticator(scanner db.Scanner, a *mfa.TOTPAuthenticator) error {
	var activatedAt pq.NullTime
	err := scanner.Scan(
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
		return err
	}
	if activatedAt.Valid {
		a.ActivatedAt = &activatedAt.Time
	}
	return nil
}

func (s *storeImpl) scanOOBAuthenticator(scanner db.Scanner, a *mfa.OOBAuthenticator) error {
	var activatedAt pq.NullTime
	err := scanner.Scan(
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
		return err
	}
	if activatedAt.Valid {
		a.ActivatedAt = &activatedAt.Time
	}
	return nil
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

	sortRecoveryCodeAuthenticatorSlice(output)

	return
}

func (s *storeImpl) DeleteRecoveryCode(userID string) error {
	old, err := s.GetRecoveryCode(userID)
	if err != nil {
		return err
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
			return err
		}

		q2 := s.sqlBuilder.Tenant().
			Delete(s.sqlBuilder.FullTableName("authenticator")).
			Where("id = ANY (?)", pq.Array(ids))

		_, err = s.sqlExecutor.ExecWith(q2)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *storeImpl) GenerateRecoveryCode(userID string) ([]mfa.RecoveryCodeAuthenticator, error) {
	err := s.DeleteRecoveryCode(userID)
	if err != nil {
		return nil, err
	}

	q1 := s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
		)
	q2 := s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("authenticator_recovery_code")).
		Columns(
			"id",
			"code",
			"created_at",
			"consumed",
		)

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
		q1 = q1.Values(
			a.ID,
			a.Type,
			a.UserID,
		)
		q2 = q2.Values(
			a.ID,
			a.Code,
			a.CreatedAt,
			a.Consumed,
		)
		output = append(output, a)
	}

	_, err = s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return nil, err
	}
	_, err = s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return nil, err
	}

	sortRecoveryCodeAuthenticatorSlice(output)

	return output, nil
}

func (s *storeImpl) UpdateRecoveryCode(a *mfa.RecoveryCodeAuthenticator) error {
	q1 := s.sqlBuilder.Tenant().
		Update(s.sqlBuilder.FullTableName("authenticator_recovery_code")).
		Set("consumed", a.Consumed).
		Where("id = ?", a.ID)
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	return err
}

func (s *storeImpl) DeleteAllBearerToken(userID string) error {
	q1 := s.sqlBuilder.Tenant().
		Select("a.id").
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Where("a.user_id = ? AND a.type = ?", userID, coreAuth.AuthenticatorTypeBearerToken)

	rows1, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return err
	}
	defer rows1.Close()

	var ids []string
	for rows1.Next() {
		var id string
		err = rows1.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	return s.deleteBearerTokenByIDs(ids)
}

func (s *storeImpl) DeleteExpiredBearerToken(userID string) error {
	now := s.timeProvider.NowUTC()
	q1 := s.sqlBuilder.Tenant().
		Select("a.id").
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_bearer_token"),
			"abt",
			"a.id = abt.id",
		).
		Where("a.user_id = ? AND abt.expire_at < ?", userID, now)

	rows1, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return err
	}
	defer rows1.Close()

	var ids []string
	for rows1.Next() {
		var id string
		err = rows1.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	return s.deleteBearerTokenByIDs(ids)
}

func (s *storeImpl) deleteBearerTokenByParentIDs(parentIDs []string) error {
	if len(parentIDs) <= 0 {
		return nil
	}

	q1 := s.sqlBuilder.Tenant().
		Select(
			"a.id",
		).
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_bearer_token"),
			"abt",
			"a.id = abt.id",
		).
		Where("abt.parent_id = ANY (?)", pq.Array(parentIDs))

	rows1, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return err
	}
	defer rows1.Close()

	var ids []string
	for rows1.Next() {
		var id string
		err = rows1.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	return s.deleteBearerTokenByIDs(ids)
}

func (s *storeImpl) deleteBearerTokenByIDs(ids []string) error {
	if len(ids) <= 0 {
		return nil
	}
	q2 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator_bearer_token")).
		Where("id = ANY (?)", pq.Array(ids))
	_, err := s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return err
	}

	q3 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator")).
		Where("id = ANY (?)", pq.Array(ids))

	_, err = s.sqlExecutor.ExecWith(q3)
	if err != nil {
		return err
	}
	return nil
}

func (s *storeImpl) CreateBearerToken(a *mfa.BearerTokenAuthenticator) error {
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
		Insert(s.sqlBuilder.FullTableName("authenticator_bearer_token")).
		Columns(
			"id",
			"parent_id",
			"token",
			"created_at",
			"expire_at",
		).
		Values(
			a.ID,
			a.ParentID,
			a.Token,
			a.CreatedAt,
			a.ExpireAt,
		)
	_, err = s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) GetBearerTokenByToken(userID string, token string) (*mfa.BearerTokenAuthenticator, error) {
	q1 := s.sqlBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"a.type",
			"abt.parent_id",
			"abt.token",
			"abt.created_at",
			"abt.expire_at",
		).
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_bearer_token"),
			"abt",
			"a.id = abt.id",
		).
		// SECURITY(louis): Ideally we should compare the bearer token in constant time.
		// However, it requires us to fetch all bearer tokens. The number can be unbound
		// because we do not limit the number of the bearer tokens.
		Where("a.user_id = ? AND abt.token = ?", userID, token)

	row, err := s.sqlExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	var a mfa.BearerTokenAuthenticator
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.Type,
		&a.ParentID,
		&a.Token,
		&a.CreatedAt,
		&a.ExpireAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			err = mfa.ErrNoAuthenticators
		}
		return nil, err
	}
	return &a, nil
}

func (s *storeImpl) ListAuthenticators(userID string) ([]mfa.Authenticator, error) {
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
		err = s.scanTOTPAuthenticator(rows1, &a)
		if err != nil {
			return nil, err
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
		err = s.scanOOBAuthenticator(rows2, &a)
		if err != nil {
			return nil, err
		}
		oobs = append(oobs, a)
	}

	output := []mfa.Authenticator{}
	for _, a := range totps {
		output = append(output, a)
	}
	for _, a := range oobs {
		output = append(output, a)
	}

	sortAuthenticatorSlice(output)

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

func (s *storeImpl) GetTOTP(userID string, id string) (*mfa.TOTPAuthenticator, error) {
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
		Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.sqlExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	var a mfa.TOTPAuthenticator
	err = s.scanTOTPAuthenticator(row, &a)
	if err != nil {
		if err == sql.ErrNoRows {
			err = mfa.ErrNoAuthenticators
		}
		return nil, err
	}
	return &a, nil
}

func (s *storeImpl) UpdateTOTP(a *mfa.TOTPAuthenticator) error {
	q1 := s.sqlBuilder.Tenant().
		Update(s.sqlBuilder.FullTableName("authenticator_totp")).
		Set("activated", a.Activated).
		Set("activated_at", a.ActivatedAt).
		Where("id = ?", a.ID)
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	return err
}

func (s *storeImpl) DeleteTOTP(a *mfa.TOTPAuthenticator) error {
	return s.deleteTOTPByIDs([]string{a.ID})
}

func (s *storeImpl) deleteTOTPByIDs(ids []string) error {
	if len(ids) <= 0 {
		return nil
	}

	err := s.deleteBearerTokenByParentIDs(ids)
	if err != nil {
		return err
	}

	q2 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator_totp")).
		Where("id = ANY (?)", pq.Array(ids))
	r2, err := s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return err
	}
	count, err := r2.RowsAffected()
	if err != nil {
		return err
	}
	if int(count) != len(ids) {
		return mfa.ErrNoAuthenticators
	}

	q3 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator")).
		Where("id = ANY (?)", pq.Array(ids))

	r3, err := s.sqlExecutor.ExecWith(q3)
	if err != nil {
		return err
	}
	count, err = r3.RowsAffected()
	if err != nil {
		return err
	}
	if int(count) != len(ids) {
		return mfa.ErrNoAuthenticators
	}

	return nil
}

func (s *storeImpl) DeleteInactiveTOTP(userID string) error {
	q1 := s.sqlBuilder.Tenant().
		Select("a.id").
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_totp"),
			"at",
			"a.id = at.id",
		).
		Where("a.user_id = ? AND at.activated = FALSE", userID)

	rows1, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return err
	}
	defer rows1.Close()

	var ids []string
	for rows1.Next() {
		var id string
		err = rows1.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	return s.deleteTOTPByIDs(ids)
}

func (s *storeImpl) GetOnlyInactiveTOTP(userID string) (*mfa.TOTPAuthenticator, error) {
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
		Where("a.user_id = ? AND at.activated = FALSE", userID)

	row, err := s.sqlExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	var a mfa.TOTPAuthenticator
	err = s.scanTOTPAuthenticator(row, &a)
	if err != nil {
		if err == sql.ErrNoRows {
			err = mfa.ErrNoAuthenticators
		}
		return nil, err
	}
	return &a, nil
}

func (s *storeImpl) CreateOOB(a *mfa.OOBAuthenticator) error {
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
		Insert(s.sqlBuilder.FullTableName("authenticator_oob")).
		Columns(
			"id",
			"activated",
			"created_at",
			"activated_at",
			"channel",
			"phone",
			"email",
		).
		Values(
			a.ID,
			a.Activated,
			a.CreatedAt,
			a.ActivatedAt,
			a.Channel,
			a.Phone,
			a.Email,
		)
	_, err = s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) GetOOB(userID string, id string) (*mfa.OOBAuthenticator, error) {
	q1 := s.sqlBuilder.Tenant().
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
		Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.sqlExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	var a mfa.OOBAuthenticator
	err = s.scanOOBAuthenticator(row, &a)
	if err != nil {
		if err == sql.ErrNoRows {
			err = mfa.ErrNoAuthenticators
		}
		return nil, err
	}
	return &a, nil
}

func (s *storeImpl) GetOnlyInactiveOOB(userID string) (*mfa.OOBAuthenticator, error) {
	q1 := s.sqlBuilder.Tenant().
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
		Where("a.user_id = ? AND ao.activated = FALSE", userID)

	row, err := s.sqlExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	var a mfa.OOBAuthenticator
	err = s.scanOOBAuthenticator(row, &a)
	if err != nil {
		if err == sql.ErrNoRows {
			err = mfa.ErrNoAuthenticators
		}
		return nil, err
	}
	return &a, nil
}

func (s *storeImpl) GetOOBByChannel(userID string, channel coreAuth.AuthenticatorOOBChannel, phone string, email string) (*mfa.OOBAuthenticator, error) {
	q1 := s.sqlBuilder.Tenant().
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
		)
	switch channel {
	case coreAuth.AuthenticatorOOBChannelSMS:
		q1 = q1.Where("a.user_id = ? AND ao.channel = ? AND ao.phone = ?", userID, channel, phone)
	case coreAuth.AuthenticatorOOBChannelEmail:
		q1 = q1.Where("a.user_id = ? AND ao.channel = ? AND ao.email = ?", userID, channel, email)
	default:
		panic("mfa: unknown authenticator channel")
	}

	row, err := s.sqlExecutor.QueryRowWith(q1)
	if err != nil {
		return nil, err
	}

	var a mfa.OOBAuthenticator
	err = s.scanOOBAuthenticator(row, &a)
	if err != nil {
		if err == sql.ErrNoRows {
			err = mfa.ErrNoAuthenticators
		}
		return nil, err
	}
	return &a, nil
}

func (s *storeImpl) UpdateOOB(a *mfa.OOBAuthenticator) error {
	q1 := s.sqlBuilder.Tenant().
		Update(s.sqlBuilder.FullTableName("authenticator_oob")).
		Set("activated", a.Activated).
		Set("activated_at", a.ActivatedAt).
		Where("id = ?", a.ID)
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	return err
}

func (s *storeImpl) DeleteOOB(a *mfa.OOBAuthenticator) error {
	return s.deleteOOBByIDs([]string{a.ID})
}

func (s *storeImpl) DeleteInactiveOOB(userID string, exceptID string) error {
	q1 := s.sqlBuilder.Tenant().
		Select("a.id").
		From(s.sqlBuilder.FullTableName("authenticator"), "a").
		Join(
			s.sqlBuilder.FullTableName("authenticator_oob"),
			"ao",
			"a.id = ao.id",
		).
		Where("a.user_id = ? AND ao.activated = FALSE", userID)

	rows1, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return err
	}
	defer rows1.Close()

	var ids []string
	for rows1.Next() {
		var id string
		err = rows1.Scan(&id)
		if err != nil {
			return err
		}
		if id != exceptID {
			ids = append(ids, id)
		}
	}

	return s.deleteOOBByIDs(ids)
}

func (s *storeImpl) deleteOOBByIDs(ids []string) error {
	if len(ids) <= 0 {
		return nil
	}

	err := s.deleteBearerTokenByParentIDs(ids)
	if err != nil {
		return err
	}

	err = s.deleteOOBCodeByAuthenticatorIDs(ids)
	if err != nil {
		return err
	}

	q1 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator_oob")).
		Where("id = ANY (?)", pq.Array(ids))
	r1, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	count, err := r1.RowsAffected()
	if err != nil {
		return err
	}
	if int(count) != len(ids) {
		return mfa.ErrNoAuthenticators
	}

	q2 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator")).
		Where("id = ANY (?)", pq.Array(ids))
	r2, err := s.sqlExecutor.ExecWith(q2)
	if err != nil {
		return err
	}
	count, err = r2.RowsAffected()
	if err != nil {
		return err
	}
	if int(count) != len(ids) {
		return mfa.ErrNoAuthenticators
	}

	return nil
}

func (s *storeImpl) GetValidOOBCode(userID string, t gotime.Time) ([]mfa.OOBCode, error) {
	q1 := s.sqlBuilder.Tenant().
		Select(
			"aoc.id",
			"a.user_id",
			"aoc.authenticator_id",
			"aoc.code",
			"aoc.created_at",
			"aoc.expire_at",
		).
		From(s.sqlBuilder.FullTableName("authenticator_oob_code"), "aoc").
		Join(
			s.sqlBuilder.FullTableName("authenticator"),
			"a",
			"a.id = aoc.authenticator_id",
		).
		Where("a.user_id = ? AND aoc.expire_at > ?", userID, t)
	rows, err := s.sqlExecutor.QueryWith(q1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var output []mfa.OOBCode
	for rows.Next() {
		var a mfa.OOBCode
		err = rows.Scan(
			&a.ID,
			&a.UserID,
			&a.AuthenticatorID,
			&a.Code,
			&a.CreatedAt,
			&a.ExpireAt,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, a)
	}

	return output, nil
}

func (s *storeImpl) CreateOOBCode(c *mfa.OOBCode) error {
	q1 := s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("authenticator_oob_code")).
		Columns(
			"id",
			"authenticator_id",
			"code",
			"created_at",
			"expire_at",
		).
		Values(
			c.ID,
			c.AuthenticatorID,
			c.Code,
			c.CreatedAt,
			c.ExpireAt,
		)
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	return nil
}

func (s *storeImpl) DeleteOOBCode(c *mfa.OOBCode) error {
	q1 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator_oob_code")).
		Where("id = ?", c.ID)
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}
	return nil
}

func (s *storeImpl) deleteOOBCodeByAuthenticatorIDs(authenticatorIDs []string) error {
	if len(authenticatorIDs) <= 0 {
		return nil
	}

	q1 := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("authenticator_oob_code")).
		Where("authenticator_id = ANY (?)", pq.Array(authenticatorIDs))
	_, err := s.sqlExecutor.ExecWith(q1)
	if err != nil {
		return err
	}

	return nil
}

var (
	_ mfa.Store = &storeImpl{}
)
