package authinfo

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

	// AssignRoles accepts array of roles and userID, the supplied roles will
	// be assigned to all passed in users
	AssignRoles(userIDs []string, roles []string) error
}

type StoreProvider struct {
	CanMigrate bool
}
