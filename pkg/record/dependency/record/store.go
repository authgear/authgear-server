package record

import (
	"fmt"
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type Store interface {
	// SetRecordAccess sets default record access of a specific type
	SetRecordAccess(recordType string, acl skydb.RecordACL) error

	// SetRecordDefaultAccess sets default record access of a specific type
	SetRecordDefaultAccess(recordType string, acl skydb.RecordACL) error

	// GetRecordAccess returns the record creation access of a specific type
	GetRecordAccess(recordType string) (skydb.RecordACL, error)

	// GetRecordDefaultAccess returns default record access of a specific type
	GetRecordDefaultAccess(recordType string) (skydb.RecordACL, error)

	// SetRecordFieldAccess replace field ACL setting
	SetRecordFieldAccess(acl skydb.FieldACL) (err error)

	// GetRecordFieldAccess retrieve field ACL setting
	GetRecordFieldAccess() (skydb.FieldACL, error)

	// GetAsset retrieves Asset information by its name
	GetAsset(name string, asset *skydb.Asset) error

	GetAssets(names []string) ([]skydb.Asset, error)

	// SaveAsset saves an Asset information into a container to
	// be referenced by records.
	SaveAsset(asset *skydb.Asset) error

	// RemoteColumnTypes returns a typemap of a database table.
	RemoteColumnTypes(recordType string) (skydb.RecordSchema, error)

	// Get fetches the Record identified by the supplied key and
	// writes it onto the supplied Record.
	//
	// Get returns an ErrRecordNotFound if Record identified by
	// the supplied key does not exist in the Database.
	// It also returns error if the underlying implementation
	// failed to read the Record.
	Get(id skydb.RecordID, record *skydb.Record) error
	GetByIDs(ids []skydb.RecordID, accessControlOptions *skydb.AccessControlOptions) (*skydb.Rows, error)

	// Save updates the supplied Record in the Database if Record with
	// the same key exists, else such Record is created.
	//
	// Save returns an error if the underlying implementation failed to
	// create / modify the Record.
	Save(record *skydb.Record) error

	// Delete removes the Record identified by the key in the Database.
	//
	// Delete returns an ErrRecordNotFound if the Record identified by
	// the supplied key does not exist in the Database.
	// It also returns an error if the underlying implementation
	// failed to remove the Record.
	Delete(id skydb.RecordID) error

	// Query executes the supplied query against the Database and returns
	// an Rows to iterate the results.
	Query(query *skydb.Query, accessControlOptions *skydb.AccessControlOptions) (*skydb.Rows, error)

	// QueryCount executes the supplied query against the Database and returns
	// the number of records matching the query's predicate.
	QueryCount(query *skydb.Query, accessControlOptions *skydb.AccessControlOptions) (uint64, error)

	// Extend extends the Database record schema such that a record
	// arrived subsequently with that schema can be saved
	//
	// Extend returns an bool indicating whether the schema is really extended.
	// Extend also returns an error if the specified schema conflicts with
	// existing schema in the Database
	Extend(recordType string, schema skydb.RecordSchema) (extended bool, err error)

	// RenameSchema renames a column of the Database record schema
	RenameSchema(recordType, oldColumnName, newColumnName string) error

	// DeleteSchema removes a column of the Database record schema
	DeleteSchema(recordType, columnName string) error

	// GetSchema returns the record schema of a record type
	GetSchema(recordType string) (skydb.RecordSchema, error)
}

// TraverseColumnTypes traverse the field type of a key path from database table.
func TraverseColumnTypes(store Store, recordType string, keyPath string) ([]skydb.FieldType, error) {
	fields := []skydb.FieldType{}
	components := strings.Split(keyPath, ".")
	for i, component := range components {
		field := skydb.FieldType{}
		isLast := (i == len(components)-1)

		schema, err := store.RemoteColumnTypes(recordType)
		if err != nil {
			return fields, fmt.Errorf(`record type "%s" does not exist`, recordType)
		}

		if f, ok := schema[component]; ok {
			field = f
		} else {
			return fields, fmt.Errorf(`keypath "%s" does not exist`, keyPath)
		}

		if field.Type != skydb.TypeReference && !isLast {
			return fields, fmt.Errorf(`field "%s" in keypath "%s" is not a reference`, component, keyPath)
		}

		fields = append(fields, field)

		if field.Type == skydb.TypeReference {
			recordType = field.ReferenceType
		}
	}
	return fields, nil
}
