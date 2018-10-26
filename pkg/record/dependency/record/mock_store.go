package record

import (
	"fmt"
	"reflect"
)

// Map is a string=>Record map
type Map map[string]Record

// SchemaMap is a string=>Schema map
type SchemaMap map[string]Schema

// MockStore is a naive memory implementation of record.Store.
type MockStore struct {
	AssetMap               map[string]Asset
	recordAccessMap        map[string]ACL
	recordDefaultAccessMap map[string]ACL
	fieldAccess            FieldACL
	Map                    Map
	SchemaMap              SchemaMap
}

// NewMockStore returns a new MockStore ready for use.
func NewMockStore() *MockStore {
	return &MockStore{
		AssetMap:               map[string]Asset{},
		recordAccessMap:        map[string]ACL{},
		recordDefaultAccessMap: map[string]ACL{},
		Map:       Map{},
		SchemaMap: SchemaMap{},
	}
}

// SetRecordAccess sets record creation access
func (s *MockStore) SetRecordAccess(recordType string, acl ACL) error {
	s.recordAccessMap[recordType] = acl
	return nil
}

// SetRecordDefaultAccess sets record creation access
func (s *MockStore) SetRecordDefaultAccess(recordType string, acl ACL) error {
	s.recordDefaultAccessMap[recordType] = acl
	return nil
}

// GetRecordAccess returns record creation access of a specific type
func (s *MockStore) GetRecordAccess(recordType string) (ACL, error) {
	acl, gotIt := s.recordAccessMap[recordType]
	if !gotIt {
		acl = NewACL([]ACLEntry{})
	}

	return acl, nil
}

// GetRecordDefaultAccess returns record default access of a specific type
func (s *MockStore) GetRecordDefaultAccess(recordType string) (ACL, error) {
	acl, gotIt := s.recordDefaultAccessMap[recordType]
	if !gotIt {
		return nil, nil
	}
	return acl, nil
}

// SetRecordFieldAccess sets record field access for all types
func (s *MockStore) SetRecordFieldAccess(acl FieldACL) error {
	s.fieldAccess = acl
	return nil
}

// GetRecordFieldAccess returns record field access for all types
func (s *MockStore) GetRecordFieldAccess() (FieldACL, error) {
	return s.fieldAccess, nil
}

// GetAsset is not implemented.
func (s *MockStore) GetAsset(name string, asset *Asset) error {
	panic("not implemented")
}

// SaveAsset is not implemented.
func (s *MockStore) SaveAsset(asset *Asset) error {
	panic("not implemented")
}

// GetAssets always returns empty array.
func (s *MockStore) GetAssets(names []string) ([]Asset, error) {
	assets := []Asset{}
	for _, v := range names {
		asset, ok := s.AssetMap[v]
		if ok {
			assets = append(assets, asset)
		}
	}
	return assets, nil
}

// RemoteColumnTypes returns a typemap of a database table.
func (s *MockStore) RemoteColumnTypes(recordType string) (Schema, error) {
	return s.SchemaMap[recordType], nil
}

// Get returns a Record from Map.
func (s *MockStore) Get(id ID, record *Record) error {
	r, ok := s.Map[id.String()]
	if !ok {
		return ErrRecordNotFound
	}
	*record = r
	return nil
}

// GetByIDs is not implemented.
func (s *MockStore) GetByIDs(ids []ID, accessControlOptions *AccessControlOptions) (*Rows, error) {
	panic("record: MockStore.GetByIDs not supported")
}

// Save assigns Record to Map.
func (s *MockStore) Save(record *Record) error {
	recordID := record.ID.String()

	if origRecord, ok := s.Map[recordID]; ok {
		// keep the meta-data of record, only update record.Data
		origRecordMergedCopy := origRecord.MergedCopy(record)
		record.Apply(&origRecordMergedCopy)
	}

	s.Map[recordID] = *record
	return nil
}

// Delete remove the specified key from Map.
func (s *MockStore) Delete(id ID) error {
	_, ok := s.Map[id.String()]
	if !ok {
		return ErrRecordNotFound
	}
	delete(s.Map, id.String())
	return nil
}

// Query is not implemented.
func (s *MockStore) Query(query *Query, accessControlOptions *AccessControlOptions) (*Rows, error) {
	panic("record: MockStore.Query not supported")
}

// QueryCount is not implemented.
func (s *MockStore) QueryCount(query *Query, accessControlOptions *AccessControlOptions) (uint64, error) {
	panic("record: MockStore.QueryCount not supported")
}

// Extend store the type of the field.
func (s *MockStore) Extend(recordType string, schema Schema) (bool, error) {
	if _, ok := s.SchemaMap[recordType]; ok {
		for fieldName, fieldType := range schema {
			if _, ok := s.SchemaMap[recordType][fieldName]; ok {
				ft := s.SchemaMap[recordType][fieldName]
				if !reflect.DeepEqual(ft, fieldType) {
					return false, fmt.Errorf("Wrong type")
				}
			}
			s.SchemaMap[recordType][fieldName] = fieldType
		}
	} else {
		s.SchemaMap[recordType] = schema
	}
	return true, nil
}

func (s *MockStore) RenameSchema(recordType, oldColumnName, newColumnName string) error {
	if _, ok := s.SchemaMap[recordType]; !ok {
		return fmt.Errorf("record type %s does not exist", recordType)
	}
	if _, ok := s.SchemaMap[recordType][oldColumnName]; !ok {
		return fmt.Errorf("column %s does not exist", oldColumnName)
	}
	if _, ok := s.SchemaMap[recordType][newColumnName]; ok {
		if !reflect.DeepEqual(
			s.SchemaMap[recordType][oldColumnName],
			s.SchemaMap[recordType][newColumnName],
		) {
			return fmt.Errorf("column type conflict")
		}
	}
	s.SchemaMap[recordType][newColumnName] = s.SchemaMap[recordType][oldColumnName]
	delete(s.SchemaMap[recordType], oldColumnName)

	return nil
}

func (s *MockStore) DeleteSchema(recordType, columnName string) error {
	if _, ok := s.SchemaMap[recordType]; !ok {
		return fmt.Errorf("record type %s does not exist", recordType)
	}
	if _, ok := s.SchemaMap[recordType][columnName]; !ok {
		return fmt.Errorf("column %s does not exist", columnName)
	}
	delete(s.SchemaMap[recordType], columnName)
	return nil
}

// GetSchema returns the record schema of a record type
func (s *MockStore) GetSchema(recordType string) (Schema, error) {
	if _, ok := s.SchemaMap[recordType]; !ok {
		return nil, fmt.Errorf("record type %s does not exist", recordType)
	}
	return s.SchemaMap[recordType], nil
}

// GetRecordSchemas returns a list of all existing record type
func (s *MockStore) GetRecordSchemas() (map[string]Schema, error) {
	return s.SchemaMap, nil
}

var (
	_ Store = NewMockStore()
)
