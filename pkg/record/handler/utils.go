package handler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/asset"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/recordconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type serializedError struct {
	id  string
	err skyerr.Error
}

func newSerializedError(id string, err skyerr.Error) serializedError {
	return serializedError{
		id:  id,
		err: err,
	}
}

func (s serializedError) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"_type":   "error",
		"name":    s.err.Name(),
		"code":    s.err.Code(),
		"message": s.err.Message(),
	}
	if s.id != "" {
		m["_id"] = s.id

		ss := strings.SplitN(s.id, "/", 2)
		if len(ss) == 2 {
			m["_recordType"] = ss[0]
			m["_recordID"] = ss[1]
		}
	}
	if s.err.Info() != nil {
		m["info"] = s.err.Info()
	}

	return json.Marshal(m)
}

type RecordModifyRequest struct {
	RecordStore   record.Store
	AssetStore    asset.Store
	Logger        *logrus.Entry
	TxContext     db.TxContext
	Atomic        bool
	WithMasterKey bool
	AuthInfo      *authinfo.AuthInfo
	ModifyAt      time.Time

	// Save only
	RecordsToSave []*record.Record

	// Delete Only
	RecordIDsToDelete []record.ID
}

type RecordModifyResponse struct {
	ErrMap           map[record.ID]skyerr.Error
	SavedRecords     []*record.Record
	DeletedRecordIDs []record.ID
}

type RecordFetcher struct {
	recordStore            record.Store
	withMasterKey          bool
	creationAccessCacheMap map[string]record.ACL
	defaultAccessCacheMap  map[string]record.ACL
	logger                 *logrus.Entry
}

// NewRecordFetcher provide a convenient FetchOrCreateRecord method
func NewRecordFetcher(recordStore record.Store, logger *logrus.Entry, withMasterKey bool) RecordFetcher {
	return RecordFetcher{
		recordStore:            recordStore,
		withMasterKey:          withMasterKey,
		creationAccessCacheMap: map[string]record.ACL{},
		defaultAccessCacheMap:  map[string]record.ACL{},
		logger:                 logger,
	}
}

func (f RecordFetcher) getCreationAccess(recordType string) record.ACL {
	creationAccess, creationAccessCached := f.creationAccessCacheMap[recordType]
	if creationAccessCached == false {
		var err error
		creationAccess, err = f.recordStore.GetRecordAccess(recordType)

		if err == nil && creationAccess != nil {
			f.creationAccessCacheMap[recordType] = creationAccess
		}
	}

	return creationAccess
}

func (f RecordFetcher) getDefaultAccess(recordType string) record.ACL {
	defaultAccess, defaultAccessCached := f.defaultAccessCacheMap[recordType]
	if defaultAccessCached == false {
		var err error
		defaultAccess, err = f.recordStore.GetRecordDefaultAccess(recordType)

		if err == nil && defaultAccess != nil {
			f.defaultAccessCacheMap[recordType] = defaultAccess
		}
	}

	return defaultAccess
}

func (f RecordFetcher) FetchRecord(recordID record.ID, authInfo *authinfo.AuthInfo, accessLevel record.ACLLevel) (r *record.Record, err skyerr.Error) {
	dbRecord := record.Record{}
	if dbErr := f.recordStore.Get(recordID, &dbRecord); dbErr != nil {
		if dbErr == record.ErrRecordNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "record not found")
		} else {
			f.logger.WithFields(logrus.Fields{
				"recordID": recordID,
				"err":      dbErr,
			}).Errorln("Failed to fetch record")
			err = skyerr.NewResourceFetchFailureErr("record", recordID.String())
		}
		return
	}

	r = &dbRecord
	if !f.withMasterKey && !dbRecord.Accessible(authInfo, accessLevel) {
		err = skyerr.NewError(
			skyerr.PermissionDenied,
			"no permission to perform operation",
		)
	}

	return
}

func (f RecordFetcher) FetchOrCreateRecord(recordID record.ID, authInfo *authinfo.AuthInfo) (r record.Record, created bool, err skyerr.Error) {
	fetchedRecord, err := f.FetchRecord(recordID, authInfo, record.WriteLevel)
	if err == nil {
		r = *fetchedRecord
		return
	}

	if err.Code() == skyerr.ResourceNotFound {
		allowCreation := func() bool {
			if f.withMasterKey {
				return true
			}

			creationAccess := f.getCreationAccess(recordID.Type)
			return creationAccess.Accessible(authInfo, record.CreateLevel)
		}()

		if !allowCreation {
			err = skyerr.NewError(
				skyerr.PermissionDenied,
				"no permission to create",
			)
			return
		}

		r = record.Record{}
		created = true
		err = nil
	}

	return
}

// RecordSaveHandler iterate the record to perform the following:
// 1. Query the db for original record
// 2. Execute before save hooks with original record and new record
// 3. Clean up some transport only data (sequence for example) away from record
// 4. Populate meta data and save the record (like updated_at/by)
// 5. Execute after save hooks with original record and new record
func RecordSaveHandler(req *RecordModifyRequest, resp *RecordModifyResponse) (err error) {
	records := req.RecordsToSave

	fetcher := NewRecordFetcher(req.RecordStore, req.Logger, req.WithMasterKey)
	var fieldACL record.FieldACL

	err = executeFuncInTx(req.TxContext, req.Atomic, func() (doErr error) {
		fieldACL, doErr = req.RecordStore.GetRecordFieldAccess()
		return doErr
	})

	if err != nil {
		return
	}

	// fetch records
	originalRecordMap := map[record.ID]*record.Record{}
	records = executeRecordsFunc(records, resp.ErrMap, req.TxContext, req.Atomic, func(r *record.Record) (err skyerr.Error) {
		dbRecord, created, err := fetcher.FetchOrCreateRecord(r.ID, req.AuthInfo)
		if err != nil {
			return err
		}

		now := req.ModifyAt
		if created {
			dbRecord.ID = r.ID
			dbRecord.OwnerID = req.AuthInfo.ID
			dbRecord.CreatedAt = now
			dbRecord.CreatorID = req.AuthInfo.ID
			dbRecord.UpdatedAt = now
			dbRecord.UpdaterID = req.AuthInfo.ID
		}

		if !req.WithMasterKey {
			if err = scrubRecordFieldsForWrite(
				req.AuthInfo,
				r,
				&dbRecord,
				fieldACL,
				req.Atomic,
			); err != nil {
				return
			}
		}

		if !created {
			origRecord := dbRecord.Copy()
			injectSigner(&origRecord, req.AssetStore)
			originalRecordMap[origRecord.ID] = &origRecord
		}

		dbRecord.Apply(r)
		*r = dbRecord
		r.UpdatedAt = now
		r.UpdaterID = req.AuthInfo.ID

		return
	})

	// Apply default access
	records = executeRecordsFunc(records, resp.ErrMap, req.TxContext, req.Atomic, func(r *record.Record) skyerr.Error {
		if r.ACL == nil {
			defaultACL := fetcher.getDefaultAccess(r.ID.Type)
			r.ACL = defaultACL
		}
		return nil
	})

	err = executeFuncInTx(req.TxContext, req.Atomic, func() error {
		return makeAssetsCompleteAndInjectSigner(req.RecordStore, records, req.AssetStore)
	})

	// TODO: before save hook

	// remove bogus field, they are only for schema change
	for _, r := range records {
		removeRecordFieldTypeHints(r)
	}

	// save records
	records = executeRecordsFunc(records, resp.ErrMap, req.TxContext, req.Atomic, func(r *record.Record) (err skyerr.Error) {
		var deltaRecord record.Record
		originalRecord, _ := originalRecordMap[r.ID]
		DeriveDeltaRecord(&deltaRecord, originalRecord, r)

		if dbErr := req.RecordStore.Save(&deltaRecord); dbErr != nil {
			err = skyerr.MakeError(dbErr)
		}
		*r = deltaRecord

		return
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		err = skyerr.NewError(skyerr.UnexpectedError, "atomic operation failed")
		return
	}

	err = executeFuncInTx(req.TxContext, req.Atomic, func() error {
		return makeAssetsCompleteAndInjectSigner(req.RecordStore, records, req.AssetStore)
	})

	// TODO: after save hook

	resp.SavedRecords = records

	return
}

type recordFunc func(*record.Record) skyerr.Error

// executeFuncInTx ensure provided function to be run in a transaction
// so if atomic is false, it would begin a new transaction
func executeFuncInTx(
	txContext db.TxContext,
	atomic bool,
	do func() error,
) (err error) {
	// Wrap function in transaction when inatomic
	// because when atomic, there would be a transaction wrapped across all functions
	if !atomic {
		txErr := db.WithTx(txContext, func() error {
			return do()
		})
		if txErr != nil {
			err = txErr
		}

		return
	}

	err = do()
	return
}

func executeRecordsFunc(
	recordsIn []*record.Record,
	errMap map[record.ID]skyerr.Error,
	txContext db.TxContext,
	atomic bool,
	rFunc recordFunc,
) (recordsOut []*record.Record) {
	for _, record := range recordsIn {
		doErr := executeFuncInTx(txContext, atomic, func() error {
			return rFunc(record)
		})

		if doErr != nil {
			errMap[record.ID] = skyerr.MakeError(doErr)
		} else {
			recordsOut = append(recordsOut, record)
		}
	}

	return
}

// scrubRecordFieldsForWrite checks the field ACL for write access.
// Depending on whether the request is an atomic one, this function
// will either remove the fields if the user is not allowed access if atomic
// is false, or will return an error.
func scrubRecordFieldsForWrite(authInfo *authinfo.AuthInfo, r *record.Record, origRecord *record.Record, fieldACL record.FieldACL, atomic bool) skyerr.Error {
	nonWritableFields := []string{}

	var deltaRecord record.Record
	DeriveDeltaRecord(&deltaRecord, origRecord, r)

	for key := range deltaRecord.Data {
		if fieldACL.Accessible(r.ID.Type, key, record.WriteFieldAccessMode, authInfo, origRecord) {
			continue
		}

		if atomic {
			nonWritableFields = append(nonWritableFields, key)
			continue
		}

		r.Remove(key)
	}

	if len(nonWritableFields) > 0 {
		return skyerr.NewDeniedArgument("Unable to save to some record fields because of Field ACL denied update.", nonWritableFields)
	}
	return nil
}

func injectSigner(r *record.Record, store asset.Store) {
	for _, value := range r.Data {
		switch v := value.(type) {
		case *record.Asset:
			if signer, ok := store.(asset.URLSigner); ok {
				v.Signer = signer
			} else {
				logrus.Warnf("Failed to acquire asset URLSigner, please check configuration")
			}
		}
	}
}

func makeAssetsCompleteAndInjectSigner(recordStore record.Store, records []*record.Record, store asset.Store) error {
	recordArr := []record.Record{}
	for _, v := range records {
		recordArr = append(recordArr, *v)
	}
	err := MakeAssetsComplete(recordStore, recordArr)
	if err != nil {
		return err
	}
	for _, record := range records {
		injectSigner(record, store)
	}
	return nil
}

func MakeAssetsComplete(recordStore record.Store, records []record.Record) error {
	if len(records) == 0 {
		return nil
	}

	recordType := records[0].ID.Type
	typemap, _ := recordStore.GetSchema(recordType)
	assetColumns := []string{}
	assetNames := []string{}

	for column, schema := range typemap {
		if schema.Type == record.TypeAsset {
			assetColumns = append(assetColumns, column)
		}
	}

	for _, r := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := r.Get(assetColumn).(*record.Asset); ok {
				assetNames = append(assetNames, thisAsset.Name)
			}
		}
	}

	if len(assetNames) == 0 {
		return nil
	}

	assets, err := recordStore.GetAssets(assetNames)
	if err != nil {
		return err
	}

	assetsByName := map[string]record.Asset{}
	for _, asset := range assets {
		assetsByName[asset.Name] = asset
	}
	for _, r := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := r.Get(assetColumn).(*record.Asset); ok {
				completeAsset := assetsByName[thisAsset.Name]
				r.Set(assetColumn, &completeAsset)
			}
		}
	}
	return nil
}

// DeriveDeltaRecord derive fields in delta which is either new or different from base, and
// write them in dst.
//
// It is the caller's reponsibility to ensure that base and delta identify
// the same record
func DeriveDeltaRecord(dst, base, delta *record.Record) {
	if base == nil {
		*dst = *delta
		return
	}

	dst.ID = delta.ID
	if delta.ACL != nil {
		dst.ACL = delta.ACL
	} else {
		dst.ACL = base.ACL
	}
	dst.OwnerID = delta.OwnerID
	dst.CreatedAt = delta.CreatedAt
	dst.CreatorID = delta.CreatorID
	dst.UpdatedAt = delta.UpdatedAt
	dst.UpdaterID = delta.UpdaterID

	dst.Data = map[string]interface{}{}
	for key, value := range delta.Data {
		if baseValue, ok := base.Data[key]; ok {
			// TODO(limouren): might want comparison that performs better
			if !reflect.DeepEqual(value, baseValue) {
				dst.Data[key] = value
			}
		} else {
			dst.Data[key] = value
		}
	}
}

func removeRecordFieldTypeHints(r *record.Record) {
	for k, v := range r.Data {
		switch v.(type) {
		case record.Sequence:
			delete(r.Data, k)
		case record.Unknown:
			delete(r.Data, k)
		}
	}
}

// RecordResultFilter is for processing Record into results.
//
// 1. Apply field-based acl, remove fields that are not accessible to the
//    provided authInfo
// 2. Inject asset
// 3. Return JSONRecord that is a copy of passed in Record that is ready to
//    be serialized
type RecordResultFilter struct {
	AssetStore          asset.Store
	TxContext           db.TxContext
	FieldACL            record.FieldACL
	AuthInfo            *authinfo.AuthInfo
	BypassAccessControl bool
}

// NewRecordResultFilter return a RecordResultFilter.
func NewRecordResultFilter(
	recordStore record.Store,
	assetStore asset.Store,
	authInfo *authinfo.AuthInfo,
	bypassAccessControl bool,
) (RecordResultFilter, error) {
	var (
		acl record.FieldACL
		err error
	)

	if !bypassAccessControl {
		acl, err = recordStore.GetRecordFieldAccess()
		if err != nil {
			return RecordResultFilter{}, err
		}
	}

	return RecordResultFilter{
		AssetStore:          assetStore,
		AuthInfo:            authInfo,
		FieldACL:            acl,
		BypassAccessControl: bypassAccessControl,
	}, nil
}

func (f *RecordResultFilter) JSONResult(r *record.Record) *recordconv.JSONRecord {
	if r == nil {
		return nil
	}

	recordCopy := r.Copy()
	if !f.BypassAccessControl {
		scrubRecordFieldsForRead(f.AuthInfo, &recordCopy, f.FieldACL)
	}
	injectSigner(r, f.AssetStore)
	return (*recordconv.JSONRecord)(&recordCopy)
}

// scrubRecordFieldsForRead checks the field ACL to remove the fields
// from a record.Record that the user is not allowed to read.
func scrubRecordFieldsForRead(authInfo *authinfo.AuthInfo, r *record.Record, fieldACL record.FieldACL) {
	for _, key := range r.UserKeys() {
		if !fieldACL.Accessible(r.ID.Type, key, record.ReadFieldAccessMode, authInfo, r) {
			r.Remove(key)
		}
	}
}

func ExtendRecordSchema(recordStore record.Store, logger *logrus.Entry, records []*record.Record) (bool, error) {
	recordSchemaMergerMap := map[string]schemaMerger{}
	for _, record := range records {
		recordType := record.ID.Type
		merger, ok := recordSchemaMergerMap[recordType]
		if !ok {
			merger = newSchemaMerger()
		}

		merger.Extend(deriveRecordSchema(record.Data))

		// The map hold the value of Schema Merger. After we have
		// updated the Schema Merger, we have to copy the value
		// of Schema Merger back to the map.
		recordSchemaMergerMap[recordType] = merger
	}

	extended := false
	for recordType, merger := range recordSchemaMergerMap {
		schema, err := merger.Schema()
		if err != nil {
			return false, err
		}

		schemaExtended, err := recordStore.Extend(recordType, schema)
		if err != nil {
			return false, err
		}
		if schemaExtended {
			logger.
				WithField("type", recordType).
				WithField("schema", schema).
				Info("Schema Extended")
			extended = true
		}
	}

	return extended, nil
}

type schemaMerger struct {
	finalSchema record.Schema
	err         error
}

func newSchemaMerger() schemaMerger {
	return schemaMerger{finalSchema: record.Schema{}}
}

func (m *schemaMerger) Extend(schema record.Schema) {
	if m.err != nil {
		return
	}

	for key, dataType := range schema {
		if originalType, ok := m.finalSchema[key]; ok {
			if originalType != dataType {
				m.err = fmt.Errorf("type conflict on column = %s, %#v -> %#v", key, originalType, dataType)
				return
			}
		}

		m.finalSchema[key] = dataType
	}
}

func (m schemaMerger) Schema() (record.Schema, error) {
	return m.finalSchema, m.err
}

func deriveRecordSchema(m record.Data) record.Schema {
	schema := record.Schema{}
	for key, value := range m {
		if value == nil {
			continue
		}

		fieldType, err := record.DeriveFieldType(value)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
			}).Panicf("unable to derive record schema: %s", err)
		}
		schema[key] = fieldType
	}

	return schema
}

type QueryResultFilter struct {
	RecordStore        record.Store
	Query              record.Query
	EagerRecords       map[string]map[string]*record.Record
	RecordResultFilter RecordResultFilter
}

func (f *QueryResultFilter) JSONResult(r *record.Record) *recordconv.JSONRecord {
	if r == nil {
		return nil
	}

	recordCopy := r.Copy()
	for transientKey, transientExpression := range f.Query.ComputedKeys {
		if transientExpression.Type != record.KeyPath {
			continue
		}

		keyPath := transientExpression.Value.(string)
		ref := getReferenceWithKeyPath(f.RecordStore, &recordCopy, keyPath)
		var transientValue interface{}
		eagerRecord := f.EagerRecords[keyPath][ref.ID.Key]
		if eagerRecord != nil {
			transientValue = f.RecordResultFilter.JSONResult(eagerRecord)
		}

		if recordCopy.Transient == nil {
			recordCopy.Transient = map[string]interface{}{}
		}
		recordCopy.Transient[transientKey] = transientValue
	}

	return f.RecordResultFilter.JSONResult(&recordCopy)
}

// getReferenceWithKeyPath returns a reference for use in eager loading
// It handles the case where reserved attribute is a string ID instead of
// a referenced ID.
func getReferenceWithKeyPath(recordStore record.Store, r *record.Record, keyPath string) record.Reference {
	valueAtKeyPath := r.Get(keyPath)
	if valueAtKeyPath == nil {
		return record.NewEmptyReference()
	}

	if ref, ok := valueAtKeyPath.(record.Reference); ok {
		return ref
	}

	// If the value at key path is not a reference, it could be a string
	// ID of a user record.
	switch keyPath {
	case "_owner_id", "_created_by", "_updated_by":
		strID, ok := valueAtKeyPath.(string)
		if !ok {
			return record.NewEmptyReference()
		}
		return record.NewReference(recordStore.UserRecordType(), strID)
	default:
		return record.NewEmptyReference()
	}
}

func getRecordCount(recordStore record.Store, query *record.Query, accessControlOptions *record.AccessControlOptions, results *record.Rows) (uint64, error) {
	if results != nil {
		recordCount := results.OverallRecordCount()
		if recordCount != nil {
			return *recordCount, nil
		}
	}

	recordCount, err := recordStore.QueryCount(query, accessControlOptions)
	if err != nil {
		return 0, err
	}

	return recordCount, nil
}

func QueryResultInfo(recordStore record.Store, query *record.Query, accessControlOptions *record.AccessControlOptions, results *record.Rows) (map[string]interface{}, error) {
	resultInfo := map[string]interface{}{}
	if query.GetCount {
		recordCount, err := getRecordCount(recordStore, query, accessControlOptions, results)
		if err != nil {
			return nil, err
		}
		resultInfo["count"] = recordCount
	}
	return resultInfo, nil
}
