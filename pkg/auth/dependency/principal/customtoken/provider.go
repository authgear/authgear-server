package customtoken

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type providerImpl struct {
	sqlBuilder        db.SQLBuilder
	sqlExecutor       db.SQLExecutor
	customTokenConfig *config.CustomTokenConfiguration
}

func newProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	customTokenConfig *config.CustomTokenConfiguration,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:        builder,
		sqlExecutor:       executor,
		customTokenConfig: customTokenConfig,
	}
}

func NewProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	customTokenConfig *config.CustomTokenConfiguration,
) Provider {
	return newProvider(builder, executor, customTokenConfig)
}

func (p *providerImpl) scan(scanner db.Scanner, principal *Principal) error {
	var rawProfileBytes []byte
	var claimsValueBytes []byte
	err := scanner.Scan(
		&principal.ID,
		&principal.UserID,
		&principal.TokenPrincipalID,
		&rawProfileBytes,
		&claimsValueBytes,
	)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawProfileBytes, &principal.RawProfile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(claimsValueBytes, &principal.ClaimsValue)
	if err != nil {
		return err
	}

	return nil
}

func (p *providerImpl) Decode(tokenString string) (claims SSOCustomTokenClaims, err error) {
	_, err = jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected JWT alg")
			}
			return []byte(p.customTokenConfig.Secret), nil
		},
	)

	return
}

func (p *providerImpl) CreatePrincipal(principal *Principal) (err error) {
	// Create principal
	builder := p.sqlBuilder.Tenant().
		Insert(p.sqlBuilder.FullTableName("principal")).
		Columns(
			"id",
			"provider",
			"user_id",
		).
		Values(
			principal.ID,
			coreAuth.PrincipalTypeCustomToken,
			principal.UserID,
		)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create principal")
	}

	rawProfileBytes, err := json.Marshal(principal.RawProfile)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create principal")
	}

	claimsBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create principal")
	}

	builder = p.sqlBuilder.Tenant().
		Insert(p.sqlBuilder.FullTableName("provider_custom_token")).
		Columns(
			"principal_id",
			"raw_profile",
			"token_principal_id",
			"claims",
		).
		Values(
			principal.ID,
			rawProfileBytes,
			principal.TokenPrincipalID,
			claimsBytes,
		)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create principal")
	}

	return
}

func (p *providerImpl) UpdatePrincipal(pp *Principal) (err error) {
	rawProfileBytes, err := json.Marshal(pp.RawProfile)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update principal")
	}

	claimsBytes, err := json.Marshal(pp.ClaimsValue)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update principal")
	}

	builder := p.sqlBuilder.Tenant().
		Update(p.sqlBuilder.FullTableName("provider_custom_token")).
		Set("raw_profile", rawProfileBytes).
		Set("claims", claimsBytes).
		Where("principal_id = ?", pp.ID)

	result, err := p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update principal")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update principal")
	}
	if rowsAffected == 0 {
		return principal.ErrNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("customtoken: want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (p *providerImpl) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"ct.token_principal_id",
			"ct.raw_profile",
			"ct.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_custom_token"), "ct", "p.id = ct.principal_id").
		Where("ct.token_principal_id = ?", tokenPrincipalID)

	scanner, err := p.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get principal by token ID")
	}

	var pp Principal
	err = p.scan(scanner, &pp)
	if err == sql.ErrNoRows {
		return nil, principal.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &pp, nil
}

func (p *providerImpl) ID() string {
	return string(coreAuth.PrincipalTypeCustomToken)
}

func (p *providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"ct.token_principal_id",
			"ct.raw_profile",
			"ct.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_custom_token"), "ct", "p.id = ct.principal_id").
		Where("p.id = ?", principalID)

	scanner, err := p.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get principal by ID")
	}

	var pp Principal
	err = p.scan(scanner, &pp)
	if err == sql.ErrNoRows {
		return nil, principal.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &pp, nil
}

func (p *providerImpl) ListPrincipalsByUserID(userID string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"ct.token_principal_id",
			"ct.raw_profile",
			"ct.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_custom_token"), "ct", "p.id = ct.principal_id").
		Where("p.user_id = ?", userID)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get principal by user ID")
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{}
		err = p.scan(rows, &principal)
		if err != nil {
			return nil, errors.HandledWithMessage(err, "failed to get principal by user ID")
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) ListPrincipalsByClaim(claimName string, claimValue string) (principals []principal.Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"ct.token_principal_id",
			"ct.raw_profile",
			"ct.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_custom_token"), "ct", "p.id = ct.principal_id").
		Where("(ct.claims #>> ?) = ?", pq.Array([]string{claimName}), claimValue)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get principal by claim")
	}
	defer rows.Close()

	for rows.Next() {
		principal := Principal{}
		err = p.scan(rows, &principal)
		if err != nil {
			return nil, errors.HandledWithMessage(err, "failed to get principal by claim")
		}
		principals = append(principals, &principal)
	}

	return
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)
