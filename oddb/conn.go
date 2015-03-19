package oddb

import (
	"errors"
)

// ErrUserDuplicated is returned by Conn.CreateUser when
// the UserInfo to be created has the same ID in the current container
var ErrUserDuplicated = errors.New("oddb: duplicated UserInfo ID")

// ErrUserNotFound is returned by Conn.GetUser, Conn.UpdateUser and
// Conn.DeleteUser when the UserInfo's ID is not found
// in the current container
var ErrUserNotFound = errors.New("oddb: UserInfo ID not found")

// ErrDeviceNotFound is returned by Conn.GetDevice when the supplied
// device ID is not found in the current container
var ErrDeviceNotFound = errors.New("oddb: Device ID not found")

// DBHookFunc specifies the interface of a database hook function
type DBHookFunc func(Database, *Record, RecordHookEvent)

// Conn encapsulates the interface of an Ourd connection to a container.
type Conn interface {
	// CRUD of UserInfo, smell like a bad design to attach these onto
	// a Conn, but looks very convenient to user.

	// CreateUser creates a new UserInfo in the container
	// this Conn associated to.
	CreateUser(userinfo *UserInfo) error

	// GetUser fetches the UserInfo with supplied ID in the container and
	// fills in the supplied UserInfo with the result.
	//
	// GetUser returns ErrUserNotFound if no UserInfo exists
	// for the supplied ID.
	GetUser(id string, userinfo *UserInfo) error

	// UpdateUser updates an existing UserInfo matched by the ID field.
	//
	// UpdateUser returns ErrUserNotFound if such UserInfo does not
	// exist in the container.
	UpdateUser(userinfo *UserInfo) error

	// DeleteUser removes UserInfo with the supplied ID in the container.
	//
	// DeleteUser returns ErrUserNotFound if such UserInfo does not
	// exist in the container.
	DeleteUser(id string) error

	GetDevice(id string, device *Device) error
	SaveDevice(device *Device) error
	DeleteDevice(id string) error

	PublicDB() Database
	PrivateDB(userKey string) Database

	AddDBRecordHook(hook DBHookFunc)

	Close() error
}
