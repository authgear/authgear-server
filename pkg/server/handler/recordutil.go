package handler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func injectSigner(record *skydb.Record, store asset.Store) {
	for _, value := range record.Data {
		switch v := value.(type) {
		case *skydb.Asset:
			if signer, ok := store.(asset.URLSigner); ok {
				v.Signer = signer
			} else {
				log.Warnf("Failed to acquire asset URLSigner, please check configuration")
			}
		}
	}
}

// scrubRecordFieldsForRead checks the field ACL to remove the fields
// from a skydb.Record that the user is not allowed to read.
func scrubRecordFieldsForRead(authInfo *skydb.AuthInfo, record *skydb.Record, fieldACL skydb.FieldACL) {
	for _, key := range record.UserKeys() {
		if !fieldACL.Accessible(record.ID.Type, key, skydb.ReadFieldAccessMode, authInfo, record) {
			record.Remove(key)
		}
	}
}

// scrubRecordFieldsForWrite checks the field ACL for write access.
// Depending on whether the request is an atomic one, this function
// will either remove the fields if the user is not allowed access if atomic
// is false, or will return an error.
func scrubRecordFieldsForWrite(authInfo *skydb.AuthInfo, record *skydb.Record, origRecord *skydb.Record, fieldACL skydb.FieldACL, atomic bool) skyerr.Error {
	nonWritableFields := []string{}

	var deltaRecord skydb.Record
	deriveDeltaRecord(&deltaRecord, origRecord, record)

	for key := range deltaRecord.Data {
		if fieldACL.Accessible(record.ID.Type, key, skydb.WriteFieldAccessMode, authInfo, origRecord) {
			continue
		}

		if atomic {
			nonWritableFields = append(nonWritableFields, key)
			continue
		}

		record.Remove(key)
	}

	if len(nonWritableFields) > 0 {
		return skyerr.NewDeniedArgument("Unable to save to some record fields because of Field ACL denied update.", nonWritableFields)
	}
	return nil
}

type recordModifyFunc func(*recordModifyRequest, *recordModifyResponse) skyerr.Error

func atomicModifyFunc(req *recordModifyRequest, resp *recordModifyResponse, mFunc recordModifyFunc) recordModifyFunc {
	return func(req *recordModifyRequest, resp *recordModifyResponse) (err skyerr.Error) {
		txDB, ok := req.Db.(skydb.TxDatabase)
		if !ok {
			err = skyerr.NewError(skyerr.NotSupported, "database impl does not support transaction")
			return
		}

		txErr := withTransaction(txDB, func() error {
			return mFunc(req, resp)
		})

		if len(resp.ErrMap) > 0 {
			info := map[string]interface{}{}
			for recordID, err := range resp.ErrMap {
				info[recordID.String()] = err
			}

			return skyerr.NewErrorWithInfo(skyerr.AtomicOperationFailure,
				"Atomic Operation rolled back due to one or more errors",
				info)
		} else if txErr != nil {
			err = skyerr.NewErrorWithInfo(skyerr.AtomicOperationFailure,
				"Atomic Operation rolled back due to an error",
				map[string]interface{}{"innerError": txErr})

		}
		return
	}
}

func withTransaction(txDB skydb.TxDatabase, do func() error) (err error) {
	err = txDB.Begin()
	if err != nil {
		return
	}

	err = do()
	if err != nil {
		if rbErr := txDB.Rollback(); rbErr != nil {
			log.Errorf("Failed to rollback: %v", rbErr)
		}

	} else {
		err = txDB.Commit()
	}

	return
}

type recordModifyRequest struct {
	Db            skydb.Database
	Conn          skydb.Conn
	AssetStore    asset.Store
	HookRegistry  *hook.Registry
	Atomic        bool
	WithMasterKey bool
	Context       context.Context
	AuthInfo      *skydb.AuthInfo

	// Save only
	RecordsToSave []*skydb.Record

	// Delete Only
	RecordIDsToDelete []skydb.RecordID
}

type recordModifyResponse struct {
	ErrMap           map[skydb.RecordID]skyerr.Error
	SchemaUpdated    bool
	SavedRecords     []*skydb.Record
	DeletedRecordIDs []skydb.RecordID
}

type recordFetcher struct {
	db                     skydb.Database
	conn                   skydb.Conn
	withMasterKey          bool
	creationAccessCacheMap map[string]skydb.RecordACL
	defaultAccessCacheMap  map[string]skydb.RecordACL
}

func newRecordFetcher(db skydb.Database, conn skydb.Conn, withMasterKey bool) recordFetcher {
	return recordFetcher{
		db:                     db,
		conn:                   conn,
		withMasterKey:          withMasterKey,
		creationAccessCacheMap: map[string]skydb.RecordACL{},
		defaultAccessCacheMap:  map[string]skydb.RecordACL{},
	}
}

func (f recordFetcher) getCreationAccess(recordType string) skydb.RecordACL {
	creationAccess, creationAccessCached := f.creationAccessCacheMap[recordType]
	if creationAccessCached == false {
		var err error
		creationAccess, err = f.conn.GetRecordAccess(recordType)

		if err == nil && creationAccess != nil {
			f.creationAccessCacheMap[recordType] = creationAccess
		}
	}

	return creationAccess
}

func (f recordFetcher) getDefaultAccess(recordType string) skydb.RecordACL {
	defaultAccess, defaultAccessCached := f.defaultAccessCacheMap[recordType]
	if defaultAccessCached == false {
		var err error
		defaultAccess, err = f.conn.GetRecordDefaultAccess(recordType)

		if err == nil && defaultAccess != nil {
			f.defaultAccessCacheMap[recordType] = defaultAccess
		}
	}

	return defaultAccess
}

func (f recordFetcher) fetchRecord(recordID skydb.RecordID, authInfo *skydb.AuthInfo, accessLevel skydb.RecordACLLevel) (record *skydb.Record, err skyerr.Error) {
	dbRecord := skydb.Record{}
	if dbErr := f.db.Get(recordID, &dbRecord); dbErr != nil {
		if dbErr == skydb.ErrRecordNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "record not found")
		} else {
			log.WithFields(logrus.Fields{
				"recordID": recordID,
				"err":      dbErr,
			}).Errorln("Failed to fetch record")
			err = skyerr.NewResourceFetchFailureErr("record", recordID.String())
		}
		return
	}

	record = &dbRecord
	if !f.withMasterKey && !dbRecord.Accessible(authInfo, accessLevel) {
		err = skyerr.NewError(
			skyerr.PermissionDenied,
			"no permission to perform operation",
		)
	}

	return
}

func (f recordFetcher) fetchOrCreateRecord(recordID skydb.RecordID, authInfo *skydb.AuthInfo) (record *skydb.Record, created bool, err skyerr.Error) {
	record, err = f.fetchRecord(recordID, authInfo, skydb.WriteLevel)
	if err == nil {
		return
	}

	if err.Code() == skyerr.ResourceNotFound {
		allowCreation := func() bool {
			if f.withMasterKey {
				return true
			}

			creationAccess := f.getCreationAccess(recordID.Type)
			return creationAccess.Accessible(authInfo, skydb.CreateLevel)
		}

		if !allowCreation() {
			err = skyerr.NewError(
				skyerr.PermissionDenied,
				"no permission to create",
			)
			return
		}

		record = &skydb.Record{}
		created = true
		err = nil
	}

	return
}

func removeRecordFieldTypeHints(r *skydb.Record) {
	for k, v := range r.Data {
		switch v.(type) {
		case skydb.Sequence:
			delete(r.Data, k)
		case skydb.Unknown:
			delete(r.Data, k)
		}
	}
}

// recordSaveHandler iterate the record to perform the following:
// 1. Query the db for original record
// 2. Execute before save hooks with original record and new record
// 3. Clean up some transport only data (sequence for example) away from record
// 4. Populate meta data and save the record (like updated_at/by)
// 5. Execute after save hooks with original record and new record
func recordSaveHandler(req *recordModifyRequest, resp *recordModifyResponse) skyerr.Error {
	db := req.Db
	records := req.RecordsToSave

	fetcher := newRecordFetcher(db, req.Conn, req.WithMasterKey)
	fieldACL, err := req.Conn.GetRecordFieldAccess()
	if err != nil {
		return skyerr.MakeError(err)
	}

	// fetch records
	originalRecordMap := map[skydb.RecordID]*skydb.Record{}
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		dbRecord, created, err := fetcher.fetchOrCreateRecord(record.ID, req.AuthInfo)
		if err != nil {
			return err
		}

		if dbRecord == nil {
			panic("unable to fetch record")
		}

		if !req.WithMasterKey {
			if err = scrubRecordFieldsForWrite(
				req.AuthInfo,
				record,
				dbRecord,
				fieldACL,
				req.Atomic,
			); err != nil {
				return
			}
		}

		now := timeNow()
		if created {
			dbRecord.ID = record.ID
			dbRecord.DatabaseID = db.ID()
			dbRecord.OwnerID = req.AuthInfo.ID
			dbRecord.CreatedAt = now
			dbRecord.CreatorID = req.AuthInfo.ID
			dbRecord.UpdatedAt = now
			dbRecord.UpdaterID = req.AuthInfo.ID
		}

		if !created {
			origRecord := dbRecord.Copy()
			injectSigner(origRecord, req.AssetStore)
			originalRecordMap[origRecord.ID] = origRecord
		}

		dbRecord.Apply(record)
		*record = *dbRecord
		record.UpdatedAt = now
		record.UpdaterID = req.AuthInfo.ID

		return
	})

	// Apply default access
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) skyerr.Error {
		if record.ACL == nil {
			defaultACL := fetcher.getDefaultAccess(record.ID.Type)
			record.ACL = defaultACL
		}
		return nil
	})

	// execute before save hooks
	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			originalRecord, _ := originalRecordMap[record.ID]
			err = req.HookRegistry.ExecuteHooks(req.Context, hook.BeforeSave, record, originalRecord)
			return
		})
	}

	// derive and extend record schema
	schemaExtended, err := extendRecordSchema(db, records)
	if err != nil {
		log.WithField("err", err).Errorln("failed to migrate record schema")
		if myerr, ok := err.(skyerr.Error); ok {
			return myerr
		}
		return skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate record schema")
	}

	// remove bogus field, they are only for schema change
	for _, r := range records {
		removeRecordFieldTypeHints(r)
	}

	// save records
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		var deltaRecord skydb.Record
		if originalRecord, ok := originalRecordMap[record.ID]; ok {
			deriveDeltaRecord(&deltaRecord, originalRecord, record)
		} else {
			deltaRecord = *record
		}

		if dbErr := db.Save(&deltaRecord); dbErr != nil {
			err = skyerr.MakeError(dbErr)
		}
		*record = deltaRecord

		return
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		return skyerr.NewError(skyerr.UnexpectedError, "atomic operation failed")
	}

	makeAssetsCompleteAndInjectSigner(db, req.Conn, records, req.AssetStore)

	// execute after save hooks
	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			originalRecord, _ := originalRecordMap[record.ID]
			err = req.HookRegistry.ExecuteHooks(req.Context, hook.AfterSave, record, originalRecord)
			if err != nil {
				log.Errorf("Error occurred while executing hooks: %s", err)
			}
			return
		})
	}

	resp.SavedRecords = records
	resp.SchemaUpdated = schemaExtended

	return nil
}

type recordFunc func(*skydb.Record) skyerr.Error

func executeRecordFunc(recordsIn []*skydb.Record, errMap map[skydb.RecordID]skyerr.Error, rFunc recordFunc) (recordsOut []*skydb.Record) {
	for _, record := range recordsIn {
		if err := rFunc(record); err != nil {
			errMap[record.ID] = err
		} else {
			recordsOut = append(recordsOut, record)
		}
	}

	return
}

// Derive fields in delta which is either new or different from base, and
// write them in dst.
//
// It is the caller's reponsibility to ensure that base and delta identify
// the same record
func deriveDeltaRecord(dst, base, delta *skydb.Record) {
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

func extendRecordSchema(db skydb.Database, records []*skydb.Record) (bool, error) {
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

		schemaExtended, err := db.Extend(recordType, schema)
		if err != nil {
			return false, err
		}
		if schemaExtended {
			log.
				WithField("type", recordType).
				WithField("schema", schema).
				Info("Schema Extended")
			extended = true
		}
	}

	return extended, nil
}

func recordDeleteHandler(req *recordModifyRequest, resp *recordModifyResponse) skyerr.Error {
	db := req.Db
	recordIDs := req.RecordIDsToDelete

	fetcher := newRecordFetcher(db, req.Conn, req.WithMasterKey)

	var records []*skydb.Record
	for _, recordID := range recordIDs {
		if recordID.Type == db.UserRecordType() {
			resp.ErrMap[recordID] = skyerr.NewError(skyerr.PermissionDenied, "cannot delete user record")
			continue
		}

		record, err := fetcher.fetchRecord(recordID, req.AuthInfo, skydb.WriteLevel)
		if err != nil {
			resp.ErrMap[recordID] = err
			continue
		}
		records = append(records, record)
	}

	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			err = req.HookRegistry.ExecuteHooks(req.Context, hook.BeforeDelete, record, nil)
			return
		})
	}

	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {

		if dbErr := db.Delete(record.ID); dbErr != nil {
			return skyerr.MakeError(dbErr)
		}
		return nil
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		return skyerr.NewError(skyerr.UnexpectedError, "atomic operation failed")
	}

	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			err = req.HookRegistry.ExecuteHooks(req.Context, hook.AfterDelete, record, nil)
			if err != nil {
				log.Errorf("Error occurred while executing hooks: %s", err)
			}
			return
		})
	}

	for _, record := range records {
		resp.DeletedRecordIDs = append(resp.DeletedRecordIDs, record.ID)
	}
	return nil
}

type schemaMerger struct {
	finalSchema skydb.RecordSchema
	err         error
}

func newSchemaMerger() schemaMerger {
	return schemaMerger{finalSchema: skydb.RecordSchema{}}
}

func (m *schemaMerger) Extend(schema skydb.RecordSchema) {
	if m.err != nil {
		return
	}

	for key, dataType := range schema {
		if originalType, ok := m.finalSchema[key]; ok {
			if originalType != dataType {
				m.err = fmt.Errorf("type conflict on column = %s, %s -> %s", key, originalType, dataType)
				return
			}
		}

		m.finalSchema[key] = dataType
	}
}

func (m schemaMerger) Schema() (skydb.RecordSchema, error) {
	return m.finalSchema, m.err
}

func deriveRecordSchema(m skydb.Data) skydb.RecordSchema {
	schema := skydb.RecordSchema{}
	log.Debugf("%v", m)
	for key, value := range m {
		if value == nil {
			continue
		}

		fieldType, err := skydb.DeriveFieldType(value)
		if err != nil {
			log.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
			}).Panicf("unable to derive record schema: %s", err)
		}
		schema[key] = fieldType
	}

	return schema
}

func eagerIDs(db skydb.Database, records []skydb.Record, query skydb.Query) map[string][]skydb.RecordID {
	eagers := map[string][]skydb.RecordID{}
	for _, transientExpression := range query.ComputedKeys {
		if transientExpression.Type != skydb.KeyPath {
			continue
		}
		keyPath := transientExpression.Value.(string)
		eagers[keyPath] = make([]skydb.RecordID, len(records))
	}

	for i, record := range records {
		for keyPath := range eagers {
			ref := getReferenceWithKeyPath(db, &record, keyPath)
			if ref.IsEmpty() {
				continue
			}
			eagers[keyPath][i] = ref.ID
		}
	}
	return eagers
}

// getReferenceWithKeyPath returns a reference for use in eager loading
// It handles the case where reserved attribute is a string ID instead of
// a referenced ID.
func getReferenceWithKeyPath(db skydb.Database, record *skydb.Record, keyPath string) skydb.Reference {
	valueAtKeyPath := record.Get(keyPath)
	if valueAtKeyPath == nil {
		return skydb.NewEmptyReference()
	}

	if ref, ok := valueAtKeyPath.(skydb.Reference); ok {
		return ref
	}

	// If the value at key path is not a reference, it could be a string
	// ID of a user record.
	switch keyPath {
	case "_owner_id", "_created_by", "_updated_by":
		strID, ok := valueAtKeyPath.(string)
		if !ok {
			return skydb.NewEmptyReference()
		}
		return skydb.NewReference(db.UserRecordType(), strID)
	default:
		return skydb.NewEmptyReference()
	}
}

func doQueryEager(db skydb.Database, eagersIDs map[string][]skydb.RecordID) map[string]map[string]*skydb.Record {
	eagerRecords := map[string]map[string]*skydb.Record{}

	for keyPath, ids := range eagersIDs {
		log.Debugf("Getting value for keypath %v", keyPath)
		eagerScanner, err := db.GetByIDs(ids)
		if err != nil {
			log.Debugf("No Records found in the eager load key path: %s", keyPath)
			eagerRecords[keyPath] = map[string]*skydb.Record{}
			continue
		}
		for eagerScanner.Scan() {
			er := eagerScanner.Record()
			if eagerRecords[keyPath] == nil {
				eagerRecords[keyPath] = map[string]*skydb.Record{}
			}
			eagerRecords[keyPath][er.ID.Key] = &er
		}
		eagerScanner.Close()
	}

	return eagerRecords
}

func getRecordCount(db skydb.Database, query *skydb.Query, results *skydb.Rows) (uint64, error) {
	if results != nil {
		recordCount := results.OverallRecordCount()
		if recordCount != nil {
			return *recordCount, nil
		}
	}

	recordCount, err := db.QueryCount(query)
	if err != nil {
		return 0, err
	}

	return recordCount, nil
}

func queryResultInfo(db skydb.Database, query *skydb.Query, results *skydb.Rows) (map[string]interface{}, error) {
	resultInfo := map[string]interface{}{}
	if query.GetCount {
		recordCount, err := getRecordCount(db, query, results)
		if err != nil {
			return nil, err
		}
		resultInfo["count"] = recordCount
	}
	return resultInfo, nil
}

func makeAssetsComplete(db skydb.Database, conn skydb.Conn, records []skydb.Record) error {
	if len(records) == 0 {
		return nil
	}

	recordType := records[0].ID.Type
	typemap, _ := db.GetSchema(recordType)
	assetColumns := []string{}
	assetNames := []string{}

	for column, schema := range typemap {
		if schema.Type == skydb.TypeAsset {
			assetColumns = append(assetColumns, column)
		}
	}

	for _, record := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := record.Get(assetColumn).(*skydb.Asset); ok {
				assetNames = append(assetNames, thisAsset.Name)
			}
		}
	}

	if len(assetNames) == 0 {
		return nil
	}

	assets, err := conn.GetAssets(assetNames)
	if err != nil {
		return err
	}

	assetsByName := map[string]skydb.Asset{}
	for _, asset := range assets {
		assetsByName[asset.Name] = asset
	}
	for _, record := range records {
		for _, assetColumn := range assetColumns {
			if thisAsset, ok := record.Get(assetColumn).(*skydb.Asset); ok {
				completeAsset := assetsByName[thisAsset.Name]
				record.Set(assetColumn, &completeAsset)
			}
		}
	}
	return nil
}

func makeAssetsCompleteAndInjectSigner(db skydb.Database, conn skydb.Conn, records []*skydb.Record, store asset.Store) error {
	recordArr := []skydb.Record{}
	for _, v := range records {
		recordArr = append(recordArr, *v)
	}
	err := makeAssetsComplete(db, conn, recordArr)
	if err != nil {
		return err
	}
	for _, record := range records {
		injectSigner(record, store)
	}
	return nil
}

type recordResultFilter struct {
	AssetStore          asset.Store
	FieldACL            skydb.FieldACL
	AuthInfo            *skydb.AuthInfo
	BypassAccessControl bool
}

func newRecordResultFilter(conn skydb.Conn, assetStore asset.Store, authInfo *skydb.AuthInfo, bypassAccessControl bool) (recordResultFilter, error) {
	var (
		acl skydb.FieldACL
		err error
	)

	if !bypassAccessControl {
		acl, err = conn.GetRecordFieldAccess()
		if err != nil {
			return recordResultFilter{}, err
		}
	}

	return recordResultFilter{
		AssetStore:          assetStore,
		AuthInfo:            authInfo,
		FieldACL:            acl,
		BypassAccessControl: bypassAccessControl,
	}, nil
}

func (f *recordResultFilter) JSONResult(record *skydb.Record) *skyconv.JSONRecord {
	if record == nil {
		return nil
	}

	if !f.BypassAccessControl {
		scrubRecordFieldsForRead(f.AuthInfo, record, f.FieldACL)
	}
	injectSigner(record, f.AssetStore)
	return (*skyconv.JSONRecord)(record)
}

type queryResultFilter struct {
	Database     skydb.Database
	Query        skydb.Query
	EagerRecords map[string]map[string]*skydb.Record
	recordResultFilter
}

func (f *queryResultFilter) JSONResult(record *skydb.Record) *skyconv.JSONRecord {
	if record == nil {
		return nil
	}

	for transientKey, transientExpression := range f.Query.ComputedKeys {
		if transientExpression.Type != skydb.KeyPath {
			continue
		}

		keyPath := transientExpression.Value.(string)
		ref := getReferenceWithKeyPath(f.Database, record, keyPath)
		var transientValue interface{}
		eagerRecord := f.EagerRecords[keyPath][ref.ID.Key]
		if eagerRecord != nil {
			transientValue = f.recordResultFilter.JSONResult(eagerRecord)
		}

		if record.Transient == nil {
			record.Transient = map[string]interface{}{}
		}
		record.Transient[transientKey] = transientValue
	}

	return f.recordResultFilter.JSONResult(record)
}
