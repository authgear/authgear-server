package password

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
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
	loginIDMetadataKeys    [][]string
	authDataChecker        authDataChecker
	passwordHistoryEnabled bool
	passwordHistoryStore   passwordhistory.Store
}

func newProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	loginIDMetadataKeys [][]string,
	passwordHistoryEnabled bool,
) *providerImpl {
	return &providerImpl{
		sqlBuilder:          builder,
		sqlExecutor:         executor,
		logger:              logger,
		loginIDMetadataKeys: loginIDMetadataKeys,
		authDataChecker: defaultAuthDataChecker{
			loginIDMetadataKeys: loginIDMetadataKeys,
		},
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
	loginIDMetadataKeys [][]string,
	passwordHistoryEnabled bool,
) Provider {
	return newProvider(builder, executor, logger, loginIDMetadataKeys, passwordHistoryEnabled)
}

func (p providerImpl) IsAuthDataValid(authData map[string]string) bool {
	return p.authDataChecker.isValid(authData)
}

func (p providerImpl) IsAuthDataMatching(authData map[string]string) bool {
	return p.authDataChecker.isMatching(authData)
}

func (p providerImpl) CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]string) (err error) {
	authDataList := toValidAuthDataList(p.loginIDMetadataKeys, authData)

	for _, a := range authDataList {
		principal := NewPrincipal()
		principal.UserID = authInfoID
		principal.AuthData = a
		principal.PlainPassword = password
		err = p.CreatePrincipal(principal)

		if err != nil {
			return
		}
	}

	return
}

func (p providerImpl) CreatePrincipal(principal Principal) (err error) {
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

	// Create password type provider data
	var authDataBytes []byte
	authDataBytes, err = json.Marshal(principal.AuthData)
	if err != nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(principal.PlainPassword), bcrypt.DefaultCost)
	if err != nil {
		panic("provider_password: Failed to hash password")
	}

	builder = p.sqlBuilder.Insert(p.sqlBuilder.FullTableName("provider_password")).Columns(
		"principal_id",
		"auth_data",
		"password",
	).Values(
		principal.ID,
		authDataBytes,
		hashedPassword,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	if p.passwordHistoryEnabled {
		p.passwordHistoryStore.CreatePasswordHistory(
			principal.UserID, hashedPassword, timeNow(),
		)
	}

	return
}

func (p providerImpl) GetPrincipalByAuthData(authData map[string]string, principal *Principal) (err error) {
	authDataBytes, err := json.Marshal(authData)
	if err != nil {
		return
	}
	builder := p.sqlBuilder.Select("principal_id", "password").
		From(p.sqlBuilder.FullTableName("provider_password")).
		Where(`auth_data = ?::jsonb`, authDataBytes)
	scanner := p.sqlExecutor.QueryRowWith(builder)

	err = scanner.Scan(
		&principal.ID,
		&principal.HashedPassword,
	)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return
	}

	principal.AuthData = authData

	builder = p.sqlBuilder.Select("user_id").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("id = ? AND provider = 'password'", principal.ID)
	scanner = p.sqlExecutor.QueryRowWith(builder)
	err = scanner.Scan(&principal.UserID)

	if err == sql.ErrNoRows {
		p.logger.Warnf("Missing principal for provider_password: %v", principal.ID)
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return err
	}

	return
}

func (p providerImpl) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Select("id", "user_id").
		From(p.sqlBuilder.FullTableName("principal")).
		Where("user_id = ? AND provider = 'password'", userID)
	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		if err = rows.Scan(
			&principal.ID,
			&principal.UserID,
		); err != nil {
			return nil, err
		}

		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	for _, principal := range principals {
		builder = p.sqlBuilder.Select("auth_data", "password").
			From(p.sqlBuilder.FullTableName("provider_password")).
			Where(`principal_id = ?`, principal.ID)
		scanner := p.sqlExecutor.QueryRowWith(builder)
		var authDataBytes []byte
		err = scanner.Scan(
			&authDataBytes,
			&principal.HashedPassword,
		)

		if err == sql.ErrNoRows {
			err = skydb.ErrUserNotFound
		}

		if err != nil {
			return
		}

		err = json.Unmarshal(authDataBytes, &principal.AuthData)

		if err != nil {
			return
		}
	}

	return
}

func (p providerImpl) GetPrincipalsByEmail(email string) (principals []*Principal, err error) {
	builder := p.sqlBuilder.Select("auth_data", "principal_id", "password").
		From(p.sqlBuilder.FullTableName("provider_password")).
		Where(`auth_data->>'email' = ?`, email)
	rows, err := p.sqlExecutor.QueryWith(builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var principal Principal
		if err = rows.Scan(
			&principal.AuthData,
			&principal.ID,
			&principal.HashedPassword,
		); err != nil {
			return
		}

		principals = append(principals, &principal)
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
		return
	}

	for _, principal := range principals {
		builder = p.sqlBuilder.Select("user_id").
			From(p.sqlBuilder.FullTableName("principal")).
			Where("id = ? AND provider = 'password'", principal.ID)
		scanner := p.sqlExecutor.QueryRowWith(builder)
		err = scanner.Scan(&principal.UserID)

		if err == sql.ErrNoRows {
			p.logger.Warnf("Missing principal for provider_password: %v", principal.ID)
			err = skydb.ErrUserNotFound
		}
		if err != nil {
			return nil, err
		}
	}

	return
}

func (p providerImpl) UpdatePrincipal(principal Principal) (err error) {
	// TODO: log

	// Create password type provider data
	var authDataBytes []byte
	authDataBytes, err = json.Marshal(principal.AuthData)
	if err != nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(principal.PlainPassword), bcrypt.DefaultCost)
	if err != nil {
		panic("provider_password: Failed to hash password")
	}

	builder := p.sqlBuilder.Update(p.sqlBuilder.FullTableName("provider_password")).
		Set("auth_data", authDataBytes).
		Set("password", hashedPassword).
		Where("principal_id = ?", principal.ID)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}

		return
	}

	var isPasswordChanged = !principal.IsSamePassword(principal.PlainPassword)
	principal.HashedPassword = hashedPassword

	if p.passwordHistoryEnabled && isPasswordChanged {
		err = p.passwordHistoryStore.CreatePasswordHistory(
			principal.UserID, hashedPassword, timeNow(),
		)
	}

	return
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &providerImpl{}
)
