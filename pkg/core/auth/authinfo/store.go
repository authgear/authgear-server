package authinfo

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq"
)

// Store encapsulates the interface of an Skygear Server connection to a container.
type Store interface {
	// CreateAuth creates a new AuthInfo in the container
	// this Conn associated to.
	CreateAuth(authinfo *AuthInfo) error

	// GetAuth fetches the AuthInfo with supplied ID in the container and
	// fills in the supplied AuthInfo with the result.
	//
	// GetAuth returns ErrUserNotFound if no AuthInfo exists
	// for the supplied ID.
	GetAuth(id string, authinfo *AuthInfo) error

	// UpdateAuth updates an existing AuthInfo matched by the ID field.
	//
	// UpdateAuth returns ErrUserNotFound if such AuthInfo does not
	// exist in the container.
	UpdateAuth(authinfo *AuthInfo) error

	// DeleteAuth removes AuthInfo with the supplied ID in the container.
	//
	// DeleteAuth returns ErrUserNotFound if such AuthInfo does not
	// exist in the container.
	DeleteAuth(id string) error
}

type StoreProvider struct {
	CanMigrate bool
}

func (p StoreProvider) Provide(ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	// TODO:
	// mock config
	dbConn, err := pq.Open(ctx, tConfig.AppName, skydb.RoleBasedAccess, tConfig.DBConnectionStr, skydb.DBConfig{
		CanMigrate: p.CanMigrate,
	})
	if err != nil {
		// TODO:
		// handle error properly
		panic(err)
	}

	return dbConn
}
