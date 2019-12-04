package password

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

var (
	timeNow = func() time.Time { return time.Now().UTC() }
)

type providerImpl struct {
	sqlBuilder             db.SQLBuilder
	sqlExecutor            db.SQLExecutor
	logger                 *logrus.Entry
	loginIDChecker         loginIDChecker
	realmChecker           realmChecker
	allowedRealms          []string
	passwordHistoryEnabled bool
	passwordHistoryStore   passwordhistory.Store
}

func newProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	loggerFactory logging.Factory,
	loginIDsKeys []config.LoginIDKeyConfiguration,
	loginIDTypes *config.LoginIDTypesConfiguration,
	allowedRealms []string,
	passwordHistoryEnabled bool,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      loggerFactory.NewLogger("password-provider"),
		loginIDChecker: defaultLoginIDChecker{
			loginIDsKeys: loginIDsKeys,
			loginIDTypes: loginIDTypes,
		},
		realmChecker: defaultRealmChecker{
			allowedRealms: allowedRealms,
		},
		allowedRealms:          allowedRealms,
		passwordHistoryEnabled: passwordHistoryEnabled,
		passwordHistoryStore: pqPWHistory.NewPasswordHistoryStore(
			builder, executor, loggerFactory,
		),
	}
}

func NewProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	loggerFactory logging.Factory,
	loginIDsKeys []config.LoginIDKeyConfiguration,
	loginIDTypes *config.LoginIDTypesConfiguration,
	allowedRealms []string,
	passwordHistoryEnabled bool,
) Provider {
	return newProvider(builder, executor, loggerFactory, loginIDsKeys, loginIDTypes, allowedRealms, passwordHistoryEnabled)
}

func (p *providerImpl) ValidateLoginID(loginID LoginID) error {
	return p.loginIDChecker.validateOne(loginID)
}

func (p *providerImpl) ValidateLoginIDs(loginIDs []LoginID) error {
	return p.loginIDChecker.validate(loginIDs)
}

func (p *providerImpl) CheckLoginIDKeyType(loginIDKey string, standardKey metadata.StandardKey) bool {
	return p.loginIDChecker.checkType(loginIDKey, standardKey)
}

func (p *providerImpl) IsRealmValid(realm string) bool {
	return p.realmChecker.isValid(realm)
}

func (p *providerImpl) IsDefaultAllowedRealms() bool {
	return len(p.allowedRealms) == 1 && p.allowedRealms[0] == DefaultRealm
}

func (p *providerImpl) scan(scanner db.Scanner, principal *Principal) error {
	var claimsValueBytes []byte

	err := scanner.Scan(
		&principal.ID,
		&principal.UserID,
		&principal.LoginIDKey,
		&principal.LoginID,
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

func (p *providerImpl) CreatePrincipalsByLoginID(authInfoID string, password string, loginIDs []LoginID, realm string) (principals []*Principal, err error) {
	// do not create principal when there is login ID belongs to another user.
	for _, loginID := range loginIDs {
		loginIDPrincipals, principalErr := p.GetPrincipalsByLoginID("", loginID.Value)
		if principalErr != nil && principalErr != principal.ErrNotFound {
			err = principalErr
			return
		}
		for _, principal := range loginIDPrincipals {
			if principal.UserID != authInfoID {
				err = ErrLoginIDAlreadyUsed
				return
			}
		}
	}

	for _, loginID := range loginIDs {
		principal := NewPrincipal()
		principal.UserID = authInfoID
		principal.LoginIDKey = loginID.Key
		principal.LoginID = loginID.Value
		principal.Realm = realm
		principal.deriveClaims(p.loginIDChecker)
		principal.setPassword(password)
		err = p.createPrincipal(principal)

		if err != nil {
			err = errors.HandledWithMessage(err, "failed to create principal")
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) createPrincipal(principal Principal) (err error) {
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
			coreAuth.PrincipalTypePassword,
			principal.UserID,
		)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	claimsValueBytes, err := json.Marshal(principal.ClaimsValue)
	if err != nil {
		return
	}

	builder = p.sqlBuilder.Tenant().
		Insert(p.sqlBuilder.FullTableName("provider_password")).
		Columns(
			"principal_id",
			"login_id_key",
			"login_id",
			"realm",
			"password",
			"claims",
		).
		Values(
			principal.ID,
			principal.LoginIDKey,
			principal.LoginID,
			principal.Realm,
			principal.HashedPassword,
			claimsValueBytes,
		)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	err = p.savePasswordHistory(&principal)

	return
}

func (p *providerImpl) savePasswordHistory(principal *Principal) error {
	if p.passwordHistoryEnabled {
		err := p.passwordHistoryStore.CreatePasswordHistory(
			principal.UserID, principal.HashedPassword, timeNow(),
		)
		if err != nil {
			return errors.Newf("failed to create password history: %w", err)
		}
	}
	return nil
}

func (p *providerImpl) GetPrincipalByLoginIDWithRealm(loginIDKey string, loginID string, realm string, pp *Principal) (err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"pp.login_id_key",
			"pp.login_id",
			"pp.realm",
			"pp.password",
			"pp.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_password"), "pp", "p.id = pp.principal_id").
		Where(`pp.login_id = ? AND pp.realm = ?`, loginID, realm)
	if loginIDKey != "" {
		builder = builder.Where("pp.login_id_key = ?", loginIDKey)
	}

	scanner, err := p.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by login ID & realm")
		return
	}

	err = p.scan(scanner, pp)
	if err == sql.ErrNoRows {
		err = principal.ErrNotFound
		return
	} else if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by login ID & realm")
		return
	}

	return
}

func (p *providerImpl) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"pp.login_id_key",
			"pp.login_id",
			"pp.realm",
			"pp.password",
			"pp.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_password"), "pp", "p.id = pp.principal_id").
		Where("p.user_id = ?", userID)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by user ID")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to get principal by user ID")
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) GetPrincipalsByClaim(claimName string, claimValue string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"pp.login_id_key",
			"pp.login_id",
			"pp.realm",
			"pp.password",
			"pp.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_password"), "pp", "p.id = pp.principal_id").
		Where("(pp.claims #>> ?) = ?", pq.Array([]string{claimName}), claimValue)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by claim")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to get principal by claim")
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) GetPrincipalsByLoginID(loginIDKey string, loginID string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"pp.login_id_key",
			"pp.login_id",
			"pp.realm",
			"pp.password",
			"pp.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_password"), "pp", "p.id = pp.principal_id").
		Where(`pp.login_id = ?`, loginID)
	if loginIDKey != "" {
		builder = builder.Where("pp.login_id_key = ?", loginIDKey)
	}

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by login ID")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to get principal by login ID")
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) UpdatePassword(principal *Principal, password string) (err error) {
	var isPasswordChanged = !principal.IsSamePassword(password)

	err = principal.setPassword(password)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to update password")
		return
	}

	builder := p.sqlBuilder.Tenant().
		Update(p.sqlBuilder.FullTableName("provider_password")).
		Set("password", principal.HashedPassword).
		Where("principal_id = ?", principal.ID)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to update password")
		return
	}

	if isPasswordChanged {
		err = p.savePasswordHistory(principal)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to update password")
			return
		}
	}

	return
}

func (p *providerImpl) MigratePassword(principal *Principal, password string) (err error) {
	migrated, err := principal.migratePassword(password)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to migrate password")
		return err
	}
	if !migrated {
		return
	}

	builder := p.sqlBuilder.Tenant().
		Update(p.sqlBuilder.FullTableName("provider_password")).
		Set("password", principal.HashedPassword).
		Where("principal_id = ?", principal.ID)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to migrate password")
		return
	}
	return
}

func (p *providerImpl) ID() string {
	return string(coreAuth.PrincipalTypePassword)
}

func (p *providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	builder := p.sqlBuilder.Tenant().
		Select(
			"p.id",
			"p.user_id",
			"pp.login_id_key",
			"pp.login_id",
			"pp.realm",
			"pp.password",
			"pp.claims",
		).
		From(p.sqlBuilder.FullTableName("principal"), "p").
		Join(p.sqlBuilder.FullTableName("provider_password"), "pp", "p.id = pp.principal_id").
		Where(`p.id = ?`, principalID)

	scanner, err := p.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by ID")
		return nil, err
	}

	pp := Principal{}
	err = p.scan(scanner, &pp)
	if err == sql.ErrNoRows {
		return nil, principal.ErrNotFound
	}
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get principal by ID")
		return nil, err
	}
	return &pp, nil
}

func (p *providerImpl) ListPrincipalsByClaim(claimName string, claimValue string) ([]principal.Principal, error) {
	principals, err := p.GetPrincipalsByClaim(claimName, claimValue)
	if err != nil {
		return nil, err
	}

	genericPrincipals := []principal.Principal{}
	for _, principal := range principals {
		genericPrincipals = append(genericPrincipals, principal)
	}

	return genericPrincipals, nil
}

func (p *providerImpl) ListPrincipalsByUserID(userID string) ([]principal.Principal, error) {
	principals, err := p.GetPrincipalsByUserID(userID)
	if err != nil {
		return nil, err
	}

	genericPrincipals := []principal.Principal{}
	for _, principal := range principals {
		genericPrincipals = append(genericPrincipals, principal)
	}

	return genericPrincipals, nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)
