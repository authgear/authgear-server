package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq"
)

type AuthInfoStoreProvider struct {
	CanMigrate bool
}

func (p AuthInfoStoreProvider) Provide(ctx context.Context, tConfig config.TenantConfiguration) interface{} {
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

// AuthInfoStore encapsulates the interface of an Skygear Server connection to a container.
type AuthInfoStore interface {
	// CRUD of AuthInfo, smell like a bad design to attach these onto
	// a Conn, but looks very convenient to user.

	// CreateAuth creates a new AuthInfo in the container
	// this Conn associated to.
	CreateAuth(authinfo *skydb.AuthInfo) error

	// GetAuth fetches the AuthInfo with supplied ID in the container and
	// fills in the supplied AuthInfo with the result.
	//
	// GetAuth returns ErrUserNotFound if no AuthInfo exists
	// for the supplied ID.
	GetAuth(id string, authinfo *skydb.AuthInfo) error

	// GetAuthByPrincipalID fetches the AuthInfo with supplied principal ID in the
	// container and fills in the supplied AuthInfo with the result.
	//
	// Principal ID is an ID of an authenticated principal with such
	// authentication provided by AuthProvider.
	//
	// GetAuthByPrincipalID returns ErrUserNotFound if no AuthInfo exists
	// for the supplied principal ID.
	GetAuthByPrincipalID(principalID string, authinfo *skydb.AuthInfo) error

	// UpdateAuth updates an existing AuthInfo matched by the ID field.
	//
	// UpdateAuth returns ErrUserNotFound if such AuthInfo does not
	// exist in the container.
	UpdateAuth(authinfo *skydb.AuthInfo) error

	// DeleteAuth removes AuthInfo with the supplied ID in the container.
	//
	// DeleteAuth returns ErrUserNotFound if such AuthInfo does not
	// exist in the container.
	DeleteAuth(id string) error
}
