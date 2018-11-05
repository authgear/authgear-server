package pq

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

type safeRecordStore struct {
	impl      record.Store
	txContext db.SafeTxContext
}

func NewSafeRecordStore(
	roleStore role.Store,
	canMigrate bool,
	sqlBuilder db.SQLBuilder,
	sqlExecutor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) record.Store {
	return &safeRecordStore{
		impl:      newRecordStore(roleStore, canMigrate, sqlBuilder, sqlExecutor, logger),
		txContext: txContext,
	}
}

func (s *safeRecordStore) UserRecordType() string {
	return s.impl.UserRecordType()
}

func (s *safeRecordStore) SetRecordAccess(recordType string, acl record.ACL) error {
	s.txContext.EnsureTx()
	return s.impl.SetRecordAccess(recordType, acl)
}

func (s *safeRecordStore) SetRecordDefaultAccess(recordType string, acl record.ACL) error {
	s.txContext.EnsureTx()
	return s.impl.SetRecordDefaultAccess(recordType, acl)
}

func (s *safeRecordStore) GetRecordAccess(recordType string) (record.ACL, error) {
	s.txContext.EnsureTx()
	return s.impl.GetRecordAccess(recordType)
}

func (s *safeRecordStore) GetRecordDefaultAccess(recordType string) (record.ACL, error) {
	s.txContext.EnsureTx()
	return s.impl.GetRecordDefaultAccess(recordType)
}

func (s *safeRecordStore) SetRecordFieldAccess(acl record.FieldACL) (err error) {
	s.txContext.EnsureTx()
	return s.impl.SetRecordFieldAccess(acl)
}

func (s *safeRecordStore) GetRecordFieldAccess() (record.FieldACL, error) {
	s.txContext.EnsureTx()
	return s.impl.GetRecordFieldAccess()
}

func (s *safeRecordStore) GetAsset(name string, asset *record.Asset) error {
	s.txContext.EnsureTx()
	return s.impl.GetAsset(name, asset)
}

func (s *safeRecordStore) GetAssets(names []string) ([]record.Asset, error) {
	s.txContext.EnsureTx()
	return s.impl.GetAssets(names)
}

func (s *safeRecordStore) SaveAsset(asset *record.Asset) error {
	s.txContext.EnsureTx()
	return s.impl.SaveAsset(asset)
}

func (s *safeRecordStore) RemoteColumnTypes(recordType string) (record.Schema, error) {
	s.txContext.EnsureTx()
	return s.impl.RemoteColumnTypes(recordType)
}

func (s *safeRecordStore) Get(id record.ID, record *record.Record) error {
	s.txContext.EnsureTx()
	return s.impl.Get(id, record)
}

func (s *safeRecordStore) GetByIDs(ids []record.ID, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	s.txContext.EnsureTx()
	return s.impl.GetByIDs(ids, accessControlOptions)
}

func (s *safeRecordStore) Save(record *record.Record) error {
	s.txContext.EnsureTx()
	return s.impl.Save(record)
}

func (s *safeRecordStore) Delete(id record.ID) error {
	s.txContext.EnsureTx()
	return s.impl.Delete(id)
}

func (s *safeRecordStore) Query(query *record.Query, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	s.txContext.EnsureTx()
	return s.impl.Query(query, accessControlOptions)
}

func (s *safeRecordStore) QueryCount(query *record.Query, accessControlOptions *record.AccessControlOptions) (uint64, error) {
	s.txContext.EnsureTx()
	return s.impl.QueryCount(query, accessControlOptions)
}

func (s *safeRecordStore) Extend(recordType string, schema record.Schema) (extended bool, err error) {
	s.txContext.EnsureTx()
	return s.impl.Extend(recordType, schema)
}

func (s *safeRecordStore) RenameSchema(recordType, oldColumnName, newColumnName string) error {
	s.txContext.EnsureTx()
	return s.impl.RenameSchema(recordType, oldColumnName, newColumnName)
}

func (s *safeRecordStore) DeleteSchema(recordType, columnName string) error {
	s.txContext.EnsureTx()
	return s.impl.DeleteSchema(recordType, columnName)
}

func (s *safeRecordStore) GetSchema(recordType string) (record.Schema, error) {
	s.txContext.EnsureTx()
	return s.impl.GetSchema(recordType)
}

func (s *safeRecordStore) GetRecordSchemas() (map[string]record.Schema, error) {
	s.txContext.EnsureTx()
	return s.impl.GetRecordSchemas()
}
