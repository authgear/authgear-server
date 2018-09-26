package password

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"golang.org/x/crypto/bcrypt"
)

type Provider interface {
	CreateEntry(principalID string, authData interface{}, hashedPassword string) error
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

func (p ProviderImpl) CreateEntry(
	principalID string,
	authData interface{},
	plainPassword string,
) error {
	var authDataBytes []byte
	authDataBytes, err := json.Marshal(authData)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		panic("provider_password: Failed to hash password")
	}

	builder := p.sqlBuilder.Insert(p.sqlBuilder.TableName("provider_password")).Columns(
		"principal_id",
		"auth_data",
		"password",
	).Values(
		principalID,
		authDataBytes,
		hashedPassword,
	)

	_, err = p.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
			err = skydb.ErrUserDuplicated
		}
	}

	return err
}
