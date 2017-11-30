// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package skydbtest

import (
	"fmt"
	"reflect"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// MapConn is a naive memory implementation of skydb.Conn
type MapConn struct {
	UserMap                map[string]skydb.AuthInfo
	AssetMap               map[string]skydb.Asset
	InternalPublicDB       skydb.Database
	recordAccessMap        map[string]skydb.RecordACL
	recordDefaultAccessMap map[string]skydb.RecordACL
	fieldAccess            skydb.FieldACL
	OAuthMap               map[string]skydb.OAuthInfo
	skydb.Conn
}

// NewMapConn returns a new MapConn.
func NewMapConn() *MapConn {
	return &MapConn{
		UserMap:                map[string]skydb.AuthInfo{},
		recordAccessMap:        map[string]skydb.RecordACL{},
		recordDefaultAccessMap: map[string]skydb.RecordACL{},
		fieldAccess:            skydb.FieldACL{},
		AssetMap:               map[string]skydb.Asset{},
		OAuthMap:               map[string]skydb.OAuthInfo{},
	}
}

// CreateAuth creates a AuthInfo in UserMap.
func (conn *MapConn) CreateAuth(authinfo *skydb.AuthInfo) error {
	if _, existed := conn.UserMap[authinfo.ID]; existed {
		return skydb.ErrUserDuplicated
	}

	conn.UserMap[authinfo.ID] = *authinfo
	return nil
}

// GetAuth returns a AuthInfo in UserMap.
func (conn *MapConn) GetAuth(id string, authinfo *skydb.AuthInfo) error {
	u, ok := conn.UserMap[id]
	if !ok {
		return skydb.ErrUserNotFound
	}

	*authinfo = u
	return nil
}

// GetAuthByPrincipalID returns a AuthInfo by its principalID.
func (conn *MapConn) GetAuthByPrincipalID(principalID string, authinfo *skydb.AuthInfo) error {
	for _, u := range conn.UserMap {
		if _, ok := u.ProviderInfo[principalID]; ok {
			*authinfo = u
			return nil
		}
	}

	return skydb.ErrUserNotFound
}

// UpdateAuth updates an existing AuthInfo in UserMap.
func (conn *MapConn) UpdateAuth(authinfo *skydb.AuthInfo) error {
	if _, ok := conn.UserMap[authinfo.ID]; !ok {
		return skydb.ErrUserNotFound
	}

	conn.UserMap[authinfo.ID] = *authinfo
	return nil
}

// DeleteAuth remove an existing in UserMap.
func (conn *MapConn) DeleteAuth(id string) error {
	if _, ok := conn.UserMap[id]; !ok {
		return skydb.ErrUserNotFound
	}

	delete(conn.UserMap, id)
	return nil
}

// GetAdminRoles is not implemented.
func (conn *MapConn) GetAdminRoles() ([]string, error) {
	return []string{
		"admin",
	}, nil
}

// SetAdminRoles is not implemented.
func (conn *MapConn) SetAdminRoles(roles []string) error {
	panic("not implemented")
}

// GetDefaultRoles always return user for testing
func (conn *MapConn) GetDefaultRoles() ([]string, error) {
	return []string{
		"user",
	}, nil
}

// SetDefaultRoles is not implemented.
func (conn *MapConn) SetDefaultRoles(roles []string) error {
	panic("not implemented")
}

// SetRecordAccess sets record creation access
func (conn *MapConn) SetRecordAccess(recordType string, acl skydb.RecordACL) error {
	conn.recordAccessMap[recordType] = acl
	return nil
}

// SetRecordDefaultAccess sets record creation access
func (conn *MapConn) SetRecordDefaultAccess(recordType string, acl skydb.RecordACL) error {
	conn.recordDefaultAccessMap[recordType] = acl
	return nil
}

// GetRecordAccess returns record creation access of a specific type
func (conn *MapConn) GetRecordAccess(recordType string) (skydb.RecordACL, error) {
	acl, gotIt := conn.recordAccessMap[recordType]
	if !gotIt {
		acl = skydb.NewRecordACL([]skydb.RecordACLEntry{})
	}

	return acl, nil
}

// GetRecordDefaultAccess returns record default access of a specific type
func (conn *MapConn) GetRecordDefaultAccess(recordType string) (skydb.RecordACL, error) {
	acl, gotIt := conn.recordDefaultAccessMap[recordType]
	if !gotIt {
		return nil, nil
	}
	return acl, nil
}

// SetRecordFieldAccess sets record field access for all types
func (conn *MapConn) SetRecordFieldAccess(acl skydb.FieldACL) error {
	conn.fieldAccess = acl
	return nil
}

// GetRecordFieldAccess returns record field access for all types
func (conn *MapConn) GetRecordFieldAccess() (skydb.FieldACL, error) {
	return conn.fieldAccess, nil
}

// GetAsset is not implemented.
func (conn *MapConn) GetAsset(name string, asset *skydb.Asset) error {
	panic("not implemented")
}

// SaveAsset is not implemented.
func (conn *MapConn) SaveAsset(asset *skydb.Asset) error {
	panic("not implemented")
}

// GetAssets always returns empty array.
func (conn *MapConn) GetAssets(names []string) ([]skydb.Asset, error) {
	assets := []skydb.Asset{}
	for _, v := range names {
		asset, ok := conn.AssetMap[v]
		if ok {
			assets = append(assets, asset)
		}
	}
	return assets, nil
}

// QueryRelation is not implemented.
func (conn *MapConn) QueryRelation(user string, name string, direction string, config skydb.QueryConfig) []skydb.AuthInfo {
	panic("not implemented")
}

// QueryRelationCount is not implemented.
func (conn *MapConn) QueryRelationCount(user string, name string, direction string) (uint64, error) {
	panic("not implemented")
}

// AddRelation is not implemented.
func (conn *MapConn) AddRelation(user string, name string, targetUser string) error {
	panic("not implemented")
}

// RemoveRelation is not implemented.
func (conn *MapConn) RemoveRelation(user string, name string, targetUser string) error {
	panic("not implemented")
}

// GetDevice is not implemented.
func (conn *MapConn) GetDevice(id string, device *skydb.Device) error {
	panic("not implemented")
}

// QueryDevicesByUser is not implemented.
func (conn *MapConn) QueryDevicesByUser(user string) ([]skydb.Device, error) {
	panic("not implemented")
}

// SaveDevice is not implemented.
func (conn *MapConn) SaveDevice(device *skydb.Device) error {
	panic("not implemented")
}

// DeleteDevice is not implemented.
func (conn *MapConn) DeleteDevice(id string) error {
	panic("not implemented")
}

// DeleteDevicesByToken is not implemented.
func (conn *MapConn) DeleteDevicesByToken(token string, t time.Time) error {
	panic("not implemented")
}

// DeleteEmptyDevicesByTime is not implemented.
func (conn *MapConn) DeleteEmptyDevicesByTime(t time.Time) error {
	panic("not implemented")
}

// PublicDB is not implemented.
func (conn *MapConn) PublicDB() skydb.Database {
	return conn.InternalPublicDB
}

// PrivateDB is not implemented.
func (conn *MapConn) PrivateDB(userKey string) skydb.Database {
	panic("not implemented")
}

// Subscribe is not implemented.
func (conn *MapConn) Subscribe(recordEventChan chan skydb.RecordEvent) error {
	panic("not implemented")
}

func (conn *MapConn) EnsureAuthRecordKeysValid(authRecordKeys [][]string) error {
	return nil
}

func (conn *MapConn) getOAuthKey(provider, principalID string) (key string) {
	return provider + "_" + principalID
}

// CreateOAuthInfo creates OAuthInfo in OAuthMap.
func (conn *MapConn) CreateOAuthInfo(oauthinfo *skydb.OAuthInfo) (err error) {
	key := conn.getOAuthKey(oauthinfo.Provider, oauthinfo.PrincipalID)
	if _, existed := conn.OAuthMap[key]; existed {
		return skydb.ErrUserDuplicated
	}

	conn.OAuthMap[key] = *oauthinfo
	return nil
}

// UpdateOAuthInfo updates an existing OAuthInfo in OAuthMap.
func (conn *MapConn) UpdateOAuthInfo(oauthinfo *skydb.OAuthInfo) (err error) {
	key := conn.getOAuthKey(oauthinfo.Provider, oauthinfo.PrincipalID)
	if _, ok := conn.OAuthMap[key]; !ok {
		return skydb.ErrUserNotFound
	}

	conn.OAuthMap[key] = *oauthinfo
	return nil
}

// GetOAuthInfo returns OAuthInfo by provider and principalID.
func (conn *MapConn) GetOAuthInfo(provider string, principalID string, oauthinfo *skydb.OAuthInfo) error {
	key := conn.getOAuthKey(provider, principalID)
	o, ok := conn.OAuthMap[key]
	if !ok {
		return skydb.ErrUserNotFound
	}

	*oauthinfo = o
	return nil
}

// GetOAuthInfoByProviderAndUserID returns OAuthInfo by provider and userID.
func (conn *MapConn) GetOAuthInfoByProviderAndUserID(provider string, userID string, oauthinfo *skydb.OAuthInfo) error {
	for _, o := range conn.OAuthMap {
		if o.Provider == provider && o.UserID == userID {
			*oauthinfo = o
			return nil
		}
	}
	return skydb.ErrUserNotFound
}

// DeleteOAuth remove an existing in OAuthMap.
func (conn *MapConn) DeleteOAuth(provider string, principalID string) error {
	key := conn.getOAuthKey(provider, principalID)
	if _, ok := conn.OAuthMap[key]; !ok {
		return skydb.ErrUserNotFound
	}

	delete(conn.OAuthMap, key)
	return nil
}

// Close does nothing.
func (conn *MapConn) Close() error {
	// do nothing
	return nil
}

// RecordMap is a string=>Record map
type RecordMap map[string]skydb.Record

// SubscriptionMap is a string=>Subscription map
type SubscriptionMap map[string]skydb.Subscription

// RecordSchemaMap is a string=>RecordSchema map
type RecordSchemaMap map[string]skydb.RecordSchema

//recordType string, acl RecordACL

// MapDB is a naive memory implementation of skydb.Database.
type MapDB struct {
	RecordMap       RecordMap
	SubscriptionMap SubscriptionMap
	RecordSchemaMap RecordSchemaMap
	DBConn          skydb.Conn
	skydb.Database
}

// NewMapDB returns a new MapDB ready for use.
func NewMapDB() *MapDB {
	return &MapDB{
		RecordMap:       RecordMap{},
		SubscriptionMap: SubscriptionMap{},
		RecordSchemaMap: RecordSchemaMap{},
		DBConn:          &MapConn{},
	}
}

func (db *MapDB) IsReadOnly() bool { return false }

func (db *MapDB) DatabaseType() skydb.DatabaseType { return skydb.PublicDatabase }

// ID returns a mock Database ID.
func (db *MapDB) ID() string {
	return ""
}

func (db *MapDB) UserRecordType() string {
	return "user"
}

// Get returns a Record from RecordMap.
func (db *MapDB) Get(id skydb.RecordID, record *skydb.Record) error {
	r, ok := db.RecordMap[id.String()]
	if !ok {
		return skydb.ErrRecordNotFound
	}
	*record = r
	return nil

}

// Save assigns Record to RecordMap.
func (db *MapDB) Save(record *skydb.Record) error {
	recordID := record.ID.String()

	if origRecord, ok := db.RecordMap[recordID]; ok {
		// keep the meta-data of record, only update record.Data
		origRecordMergedCopy := origRecord.MergedCopy(record)
		record.Apply(&origRecordMergedCopy)
	}

	db.RecordMap[recordID] = *record
	return nil
}

// Delete remove the specified key from RecordMap.
func (db *MapDB) Delete(id skydb.RecordID) error {
	_, ok := db.RecordMap[id.String()]
	if !ok {
		return skydb.ErrRecordNotFound
	}
	delete(db.RecordMap, id.String())
	return nil
}

// Query is not implemented.
func (db *MapDB) Query(query *skydb.Query) (*skydb.Rows, error) {
	panic("skydbtest: MapDB.Query not supported")
}

// Extend store the type of the field.
func (db *MapDB) Extend(recordType string, schema skydb.RecordSchema) (bool, error) {
	if _, ok := db.RecordSchemaMap[recordType]; ok {
		for fieldName, fieldType := range schema {
			if _, ok := db.RecordSchemaMap[recordType][fieldName]; ok {
				ft := db.RecordSchemaMap[recordType][fieldName]
				if !reflect.DeepEqual(ft, fieldType) {
					return false, fmt.Errorf("Wrong type")
				}
			}
			db.RecordSchemaMap[recordType][fieldName] = fieldType
		}
	} else {
		db.RecordSchemaMap[recordType] = schema
	}
	return true, nil
}

func (db *MapDB) RenameSchema(recordType, oldColumnName, newColumnName string) error {
	if _, ok := db.RecordSchemaMap[recordType]; !ok {
		return fmt.Errorf("record type %s does not exist", recordType)
	}
	if _, ok := db.RecordSchemaMap[recordType][oldColumnName]; !ok {
		return fmt.Errorf("column %s does not exist", oldColumnName)
	}
	if _, ok := db.RecordSchemaMap[recordType][newColumnName]; ok {
		if !reflect.DeepEqual(
			db.RecordSchemaMap[recordType][oldColumnName],
			db.RecordSchemaMap[recordType][newColumnName],
		) {
			return fmt.Errorf("column type conflict")
		}
	}
	db.RecordSchemaMap[recordType][newColumnName] = db.RecordSchemaMap[recordType][oldColumnName]
	delete(db.RecordSchemaMap[recordType], oldColumnName)

	return nil
}

func (db *MapDB) DeleteSchema(recordType, columnName string) error {
	if _, ok := db.RecordSchemaMap[recordType]; !ok {
		return fmt.Errorf("record type %s does not exist", recordType)
	}
	if _, ok := db.RecordSchemaMap[recordType][columnName]; !ok {
		return fmt.Errorf("column %s does not exist", columnName)
	}
	delete(db.RecordSchemaMap[recordType], columnName)
	return nil
}

// GetSchema returns the record schema of a record type
func (db *MapDB) GetSchema(recordType string) (skydb.RecordSchema, error) {
	if _, ok := db.RecordSchemaMap[recordType]; !ok {
		return nil, fmt.Errorf("record type %s does not exist", recordType)
	}
	return db.RecordSchemaMap[recordType], nil
}

// GetRecordSchemas returns a list of all existing record type
func (db *MapDB) GetRecordSchemas() (map[string]skydb.RecordSchema, error) {
	return db.RecordSchemaMap, nil
}

// GetSubscription return a Subscription from SubscriptionMap.
func (db *MapDB) GetSubscription(name string, deviceID string, subscription *skydb.Subscription) error {
	s, ok := db.SubscriptionMap[deviceID+"/"+name]
	if !ok {
		return skydb.ErrSubscriptionNotFound
	}
	*subscription = s
	return nil
}

// SaveSubscription assigns to SubscriptionMap.
func (db *MapDB) SaveSubscription(subscription *skydb.Subscription) error {
	db.SubscriptionMap[subscription.DeviceID+"/"+subscription.ID] = *subscription
	return nil
}

// DeleteSubscription deletes the specified key from SubscriptionMap.
func (db *MapDB) DeleteSubscription(name string, deviceID string) error {
	key := deviceID + "/" + name
	_, ok := db.SubscriptionMap[key]
	if !ok {
		return skydb.ErrSubscriptionNotFound
	}
	delete(db.SubscriptionMap, key)
	return nil
}

// MockTxDatabase implements and records TxDatabase's methods and delegates other
// calls to underlying Database
type MockTxDatabase struct {
	DidBegin, DidCommit, DidRollback bool
	skydb.Database
}

func NewMockTxDatabase(backingDB skydb.Database) *MockTxDatabase {
	return &MockTxDatabase{Database: backingDB}
}

func (db *MockTxDatabase) Begin() error {
	db.DidBegin = true
	return nil
}

func (db *MockTxDatabase) Commit() error {
	db.DidCommit = true
	return nil
}

func (db *MockTxDatabase) Rollback() error {
	db.DidRollback = true
	return nil
}

var _ skydb.TxDatabase = &MockTxDatabase{}

var (
	_ skydb.Conn       = NewMapConn()
	_ skydb.Database   = NewMapDB()
	_ skydb.TxDatabase = &MockTxDatabase{}
)
