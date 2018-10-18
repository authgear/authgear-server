package password

import (
	"database/sql"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"golang.org/x/crypto/bcrypt"
)

const providerPassword string = "password"

type Provider interface {
	CreatePrincipal(principal Principal) error
	GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) error
	GetPrincipalByUserID(userID string, principal *Principal) error
	UpdatePrincipal(principal Principal) error
}

type ProviderImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func NewProvider(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *ProviderImpl {
	return &ProviderImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (p ProviderImpl) CreatePrincipal(principal Principal) (err error) {
	// TODO: log

	// Create principal
	builder := p.sqlBuilder.Insert(p.sqlBuilder.TableName("principal")).Columns(
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

	builder = p.sqlBuilder.Insert(p.sqlBuilder.TableName("provider_password")).Columns(
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

	return
}

func (p ProviderImpl) GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) (err error) {
	authDataBytes, err := json.Marshal(authData)
	if err != nil {
		return
	}
	builder := p.sqlBuilder.Select("principal_id", "password").
		From(p.sqlBuilder.TableName("provider_password")).
		Where(`auth_data @> ?::jsonb`, authDataBytes)
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
		From(p.sqlBuilder.TableName("principal")).
		Where("id = ? AND provider = 'password'", principal.ID)
	scanner = p.sqlExecutor.QueryRowWith(builder)
	err = scanner.Scan(&principal.UserID)

	if err == sql.ErrNoRows {
		p.logger.Warnf("Missing principal for provider_password: %v", principal.ID)
		err = skydb.ErrUserNotFound
	}

	return
}

func (p ProviderImpl) GetPrincipalByUserID(userID string, principal *Principal) (err error) {
	builder := p.sqlBuilder.Select("id", "user_id").
		From(p.sqlBuilder.TableName("principal")).
		Where("user_id = ? AND provider = 'password'", userID)
	scanner := p.sqlExecutor.QueryRowWith(builder)
	err = scanner.Scan(
		&principal.ID,
		&principal.UserID,
	)

	if err == sql.ErrNoRows {
		err = skydb.ErrUserNotFound
	}

	if err != nil {
		return
	}

	builder = p.sqlBuilder.Select("auth_data", "password").
		From(p.sqlBuilder.TableName("provider_password")).
		Where(`principal_id = ?`, principal.ID)
	scanner = p.sqlExecutor.QueryRowWith(builder)
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

	return
}

func (p ProviderImpl) UpdatePrincipal(principal Principal) (err error) {
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

	builder := p.sqlBuilder.Update(p.sqlBuilder.TableName("provider_password")).
		Set("auth_data", authDataBytes).
		Set("password", hashedPassword).
		Where("principal_id = ?", principal.ID)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	return
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Provider = &ProviderImpl{}
)
