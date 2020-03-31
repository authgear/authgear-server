package password

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type Store interface {
	CreatePrincipal(principal *Principal) error
	DeletePrincipal(principal *Principal) error
	GetPrincipals(loginIDKey string, loginID string, realm *string) ([]*Principal, error)
	GetPrincipalByID(principalID string) (principal.Principal, error)
	GetPrincipalsByUserID(userID string) ([]*Principal, error)
	GetPrincipalsByClaim(claimName string, claimValue string) ([]*Principal, error)
	UpdatePassword(principal *Principal, password string) (err error)
}

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
}

func NewStore(builder db.SQLBuilder, executor db.SQLExecutor) Store {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
	}
}

func (s *storeImpl) CreatePrincipal(principal *Principal) (err error) {
	builder := s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("principal")).
		Columns(
			"id",
			"provider",
			"user_id",
		).
		Values(
			principal.ID,
			coreauthn.PrincipalTypePassword,
			principal.UserID,
		)

	_, err = s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	claimsValueBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return
	}

	builder = s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("provider_password")).
		Columns(
			"principal_id",
			"login_id_key",
			"login_id",
			"original_login_id",
			"unique_key",
			"realm",
			"password",
			"claims",
		).
		Values(
			principal.ID,
			principal.LoginIDKey,
			principal.LoginID,
			principal.OriginalLoginID,
			principal.UniqueKey,
			principal.Realm,
			principal.HashedPassword,
			claimsValueBytes,
		)

	_, err = s.sqlExecutor.ExecWith(builder)
	if err != nil {
		if isUniqueViolated(err) {
			err = ErrLoginIDAlreadyUsed
		}
	}

	return
}

func (s *storeImpl) DeletePrincipal(principal *Principal) error {
	builder := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("provider_password")).
		Where("principal_id = ?", principal.ID)

	_, err := s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	builder = s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("principal")).
		Where("id = ?", principal.ID)

	_, err = s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) GetPrincipals(loginIDKey string, loginID string, realm *string) (principals []*Principal, err error) {
	builder := s.selectBuilder().
		Where(`pp.login_id = ? AND pp.login_id_key = ?`, loginID, loginIDKey)
	if realm != nil {
		builder = builder.Where("pp.realm = ?", *realm)
	}

	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = s.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (s *storeImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	builder := s.selectBuilder().
		Where(`p.id = ?`, principalID)

	scanner, err := s.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by ID")
		return nil, err
	}

	pp := Principal{}
	err = s.scan(scanner, &pp)
	if err == sql.ErrNoRows {
		return nil, principal.ErrNotFound
	}
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by ID")
		return nil, err
	}
	return &pp, nil
}

func (s *storeImpl) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	builder := s.selectBuilder().
		Where("p.user_id = ?", userID)

	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by user ID")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = s.scan(rows, &principal)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to get principal by user ID")
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (s *storeImpl) GetPrincipalsByClaim(claimName string, claimValue string) (principals []*Principal, err error) {
	builder := s.selectBuilder().
		Where("(pp.claims #>> ?) = ?", pq.Array([]string{claimName}), claimValue)

	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by claim")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = s.scan(rows, &principal)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to get principal by claim")
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (s *storeImpl) UpdatePassword(principal *Principal, password string) (err error) {
	builder := s.sqlBuilder.Tenant().
		Update(s.sqlBuilder.FullTableName("provider_password")).
		Set("password", principal.HashedPassword).
		Where("principal_id = ?", principal.ID)

	_, err = s.sqlExecutor.ExecWith(builder)
	return
}

func (s *storeImpl) selectBuilder() db.SelectBuilder {
	return s.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"pp.login_id_key",
			"pp.login_id",
			"pp.original_login_id",
			"pp.unique_key",
			"pp.realm",
			"pp.password",
			"pp.claims",
		).
		From(s.sqlBuilder.FullTableName("principal"), "p").
		Join(s.sqlBuilder.FullTableName("provider_password"), "pp", "p.id = pp.principal_id")
}

func (s *storeImpl) scan(scanner db.Scanner, principal *Principal) error {
	var claimsValueBytes []byte

	err := scanner.Scan(
		&principal.ID,
		&principal.UserID,
		&principal.LoginIDKey,
		&principal.LoginID,
		&principal.OriginalLoginID,
		&principal.UniqueKey,
		&principal.Realm,
		&principal.HashedPassword,
		&claimsValueBytes,
	)
	if err != nil {
		return err
	}

	err = json.Unmarshal(claimsValueBytes, &principal.ClaimsValue)
	if err != nil {
		return err
	}

	return nil
}

func isUniqueViolated(err error) bool {
	for ; err != nil; err = errors.Unwrap(err) {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return true
		}
	}
	return false
}

var (
	_ Store = &storeImpl{}
)
