package password

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
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
	logger *logrus.Entry,
	loginIDsKeys map[string]config.LoginIDKeyConfiguration,
	allowedRealms []string,
	passwordHistoryEnabled bool,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
		loginIDChecker: defaultLoginIDChecker{
			loginIDsKeys: loginIDsKeys,
		},
		realmChecker: defaultRealmChecker{
			allowedRealms: allowedRealms,
		},
		allowedRealms:          allowedRealms,
		passwordHistoryEnabled: passwordHistoryEnabled,
		passwordHistoryStore: pqPWHistory.NewPasswordHistoryStore(
			builder, executor, logger,
		),
	}
}

func NewProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	loginIDsKeys map[string]config.LoginIDKeyConfiguration,
	allowedRealms []string,
	passwordHistoryEnabled bool,
) Provider {
	return newProvider(builder, executor, logger, loginIDsKeys, allowedRealms, passwordHistoryEnabled)
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
		if principalErr != nil && principalErr != skydb.ErrUserNotFound {
			err = principalErr
			return
		}
		for _, principal := range loginIDPrincipals {
			if principal.UserID != authInfoID {
				err = skydb.ErrUserDuplicated
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
		err = p.CreatePrincipal(principal)

		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	return
}

func (p *providerImpl) CreatePrincipal(principal Principal) (err error) {
	// TODO: log

	// Create principal
	builder := p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("principal")).Columns(
		"id",
		"provider",
		"user_id",
	).Values(
		principal.ID,
		providerPassword,
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

	builder = p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("provider_password")).Columns(
		"principal_id",
		"login_id_key",
		"login_id",
		"realm",
		"password",
		"claims",
	).Values(
		principal.ID,
		principal.LoginIDKey,
		principal.LoginID,
		principal.Realm,
		principal.HashedPassword,
		claimsValueBytes,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	if p.passwordHistoryEnabled {
		p.passwordHistoryStore.CreatePasswordHistory(
			principal.UserID, principal.HashedPassword, timeNow(),
		)
	}

	return
}

func (p *providerImpl) GetPrincipalByLoginIDWithRealm(loginIDKey string, loginID string, realm string, principal *Principal) (err error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"pp.login_id_key",
		"pp.login_id",
		"pp.realm",
		"pp.password",
		"pp.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS pp ON p.id = pp.principal_id", p.sqlBuilder.FullTableName("provider_password"))).
		Where(`pp.login_id = ? AND pp.realm = ?`, loginID, realm)
	if loginIDKey != "" {
		builder = builder.Where("pp.login_id_key = ?", loginIDKey)
	}

	scanner := p.sqlExecutor.QueryRowWith(builder)

	err = p.scan(scanner, principal)
	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}
	if err != nil {
		return
	}

	return
}

func (p *providerImpl) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"pp.login_id_key",
		"pp.login_id",
		"pp.realm",
		"pp.password",
		"pp.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS pp ON p.id = pp.principal_id", p.sqlBuilder.FullTableName("provider_password"))).
		Where("p.user_id = ?", userID)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	return
}

func (p *providerImpl) GetPrincipalsByClaim(claimName string, claimValue string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"pp.login_id_key",
		"pp.login_id",
		"pp.realm",
		"pp.password",
		"pp.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS pp ON p.id = pp.principal_id", p.sqlBuilder.FullTableName("provider_password"))).
		Where(fmt.Sprintf(`(pp.claims #>> '{%s}') = ?`, claimName), claimValue)

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	return
}

func (p *providerImpl) GetPrincipalsByLoginID(loginIDKey string, loginID string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"pp.login_id_key",
		"pp.login_id",
		"pp.realm",
		"pp.password",
		"pp.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS pp ON p.id = pp.principal_id", p.sqlBuilder.FullTableName("provider_password"))).
		Where(`pp.login_id = ?`, loginID)
	if loginIDKey != "" {
		builder = builder.Where("pp.login_id_key = ?", loginIDKey)
	}

	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		err = p.scan(rows, &principal)
		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	return
}

func (p *providerImpl) UpdatePassword(principal *Principal, password string) (err error) {
	// TODO: log

	var isPasswordChanged = !principal.IsSamePassword(password)

	err = principal.setPassword(password)
	if err != nil {
		panic("provider_password: Failed to hash password")
	}

	builder := p.sqlBuilder.Update(p.sqlBuilder.FullTableName("provider_password")).
		Set("password", principal.HashedPassword).
		Where("principal_id = ?", principal.ID)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}

		return
	}

	if p.passwordHistoryEnabled && isPasswordChanged {
		err = p.passwordHistoryStore.CreatePasswordHistory(
			principal.UserID, principal.HashedPassword, timeNow(),
		)
	}

	return
}

func (p *providerImpl) ID() string {
	return providerPassword
}

func (p *providerImpl) GetPrincipalByID(principalID string) (principal.Principal, error) {
	builder := p.sqlBuilder.Select(
		"p.id",
		"p.user_id",
		"pp.login_id_key",
		"pp.login_id",
		"pp.realm",
		"pp.password",
		"pp.claims",
	).
		From(fmt.Sprintf("%s AS p", p.sqlBuilder.FullTableName("principal"))).
		Join(fmt.Sprintf("%s AS pp ON p.id = pp.principal_id", p.sqlBuilder.FullTableName("provider_password"))).
		Where(`p.id = ?`, principalID)

	scanner := p.sqlExecutor.QueryRowWith(builder)

	principal := Principal{}
	err := p.scan(scanner, &principal)
	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &principal, nil
}

func (p *providerImpl) ListPrincipalsByClaim(claimName string, claimValue string) ([]principal.Principal, error) {
	principals, err := p.GetPrincipalsByClaim(claimName, claimValue)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			return nil, nil
		}
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
		if err == skydb.ErrUserNotFound {
			return nil, nil
		}
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
