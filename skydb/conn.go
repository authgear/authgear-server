package skydb

import (
	"errors"
	"time"
)

// ErrUserDuplicated is returned by Conn.CreateUser when
// the UserInfo to be created has the same ID/username in the current container
var ErrUserDuplicated = errors.New("skydb: duplicated UserID or Username")

// ErrUserNotFound is returned by Conn.GetUser, Conn.UpdateUser and
// Conn.DeleteUser when the UserInfo's ID is not found
// in the current container
var ErrUserNotFound = errors.New("skydb: UserInfo ID not found")

// ErrDeviceNotFound is returned by Conn.GetDevice, Conn.DeleteDevice,
// Conn.DeleteDeviceByToken and Conn.DeleteEmptyDevicesByTime, if the desired Device
// cannot be found in the current container
var ErrDeviceNotFound = errors.New("skydb: Specific device not found")

// ZeroTime represent a zero time.Time. It is used in DeleteDeviceByToken and
// DeleteEmptyDevicesByTime to signify a Delete without time constraint.
var ZeroTime = time.Time{}

// DBHookFunc specifies the interface of a database hook function
type DBHookFunc func(Database, *Record, RecordHookEvent)

// QueryConfig provides optional parameters for queries.
// result is unlimited if Limit=0
type QueryConfig struct {
	Limit  uint64
	Offset uint64
}

// Conn encapsulates the interface of an Skygear connection to a container.
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

	// GetUserByUsernameEmail fetch the UserInfo with supplied username or email,
	//
	GetUserByUsernameEmail(username string, email string, userinfo *UserInfo) error

	// GetUserByPrincipalID fetches the UserInfo with supplied principal ID in the
	// container and fills in the supplied UserInfo with the result.
	//
	// Principal ID is an ID of an authenticated principal with such
	// authentication provided by AuthProvider.
	//
	// GetUserByPrincipalID returns ErrUserNotFound if no UserInfo exists
	// for the supplied principal ID.
	GetUserByPrincipalID(principalID string, userinfo *UserInfo) error

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

	// GetAsset retrieves Asset information by its name
	GetAsset(name string, asset *Asset) error

	// SaveAsset saves an Asset information into a container to
	// be referenced by records.
	SaveAsset(asset *Asset) error

	QueryRelation(user string, name string, direction string, config QueryConfig) []UserInfo
	QueryRelationCount(user string, name string, direction string) (uint64, error)
	AddRelation(user string, name string, targetUser string) error
	RemoveRelation(user string, name string, targetUser string) error

	GetDevice(id string, device *Device) error

	// QueryDevicesByUser queries the Device database which are registered
	// by the specified user.
	QueryDevicesByUser(user string) ([]Device, error)
	SaveDevice(device *Device) error
	DeleteDevice(id string) error

	// DeleteDeviceByToken deletes device where its Token == token and
	// LastRegisteredAt < t. If t == ZeroTime, LastRegisteredAt is not considered.
	//
	// If such device does not exist, ErrDeviceNotFound is returned.
	DeleteDeviceByToken(token string, t time.Time) error

	// DeleteEmptyDevicesByTime deletes device where Token is empty and
	// LastRegisteredAt < t. If t == ZeroTime, LastRegisteredAt is not considered.
	//
	// If such device does not exist, ErrDeviceNotFound is returned.
	DeleteEmptyDevicesByTime(t time.Time) error

	PublicDB() Database
	PrivateDB(userKey string) Database

	// Subscribe registers the specified recordEventChan to receive
	// RecordEvent from the Conn implementation
	Subscribe(recordEventChan chan RecordEvent) error

	Close() error
}

// AccessModel indicates the type of access control model while db query.
type AccessModel int

const (
	RoleBaseAC AccessModel = iota + 1
	RelationBaseAC
)

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
