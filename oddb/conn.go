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

	// QueryUser queries for UserInfo matching one of the specified emails.
	QueryUser(emails []string) ([]UserInfo, error)

	// DeleteUser removes UserInfo with the supplied ID in the container.
	//
	// DeleteUser returns ErrUserNotFound if such UserInfo does not
	// exist in the container.
	DeleteUser(id string) error

	QueryRelation(user string, name string, direction string) []UserInfo
	AddRelation(user string, name string, targetUser string) error
	RemoveRelation(user string, name string, targetUser string) error

	GetDevice(id string, device *Device) error
	SaveDevice(device *Device) error
	DeleteDevice(id string) error

	PublicDB() Database
	PrivateDB(userKey string) Database

	// Subscribe registers the specified recordEventChan to receive
	// RecordEvent from the Conn implementation
	Subscribe(recordEventChan chan RecordEvent) error

	Close() error
}

// RecordHookEvent indicates the type of record event that triggered
// the hook
type RecordHookEvent int

// See the definition of RecordHookEvent
const (
	RecordCreated RecordHookEvent = iota + 1
	RecordUpdated
	RecordDeleted
)

// RecordEvent describes a change event on Record which is either
// Created, Updated or Deleted.
//
// For RecordCreated or RecordUpdated event, Record is the newly
// created / updated Record. For RecordDeleted, Record is the Record
// being deleted.
type RecordEvent struct {
	Record *Record
	Event  RecordHookEvent
}
