package handler

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
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
	UserInfo      *skydb.UserInfo

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
}

func newRecordFetcher(db skydb.Database, conn skydb.Conn, withMasterKey bool) recordFetcher {
	return recordFetcher{
		db:                     db,
		conn:                   conn,
		withMasterKey:          withMasterKey,
		creationAccessCacheMap: map[string]skydb.RecordACL{},
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

func (f recordFetcher) fetchOrCreateRecord(recordID skydb.RecordID, userInfo *skydb.UserInfo) (record *skydb.Record, err skyerr.Error) {
	dbRecord := skydb.Record{}
	if dbErr := f.db.Get(recordID, &dbRecord); dbErr != nil {
		if dbErr == skydb.ErrRecordNotFound {
			// new record
			if f.withMasterKey {
				return
			}

			creationAccess := f.getCreationAccess(recordID.Type)
			if !creationAccess.Accessible(userInfo, skydb.CreateLevel) {
				err = skyerr.NewError(
					skyerr.PermissionDenied,
					"no permission to create",
				)
			}

			return
		}
		return nil, skyerr.NewError(skyerr.UnexpectedError, dbErr.Error())
	}

	record = &dbRecord
	if !f.withMasterKey && !dbRecord.Accessible(userInfo, skydb.WriteLevel) {
		err = skyerr.NewError(
			skyerr.PermissionDenied,
			"no permission to modify",
		)
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

	// fetch records
	originalRecordMap := map[skydb.RecordID]*skydb.Record{}
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		dbRecord, err := fetcher.fetchOrCreateRecord(record.ID, req.UserInfo)

		if dbRecord == nil || err != nil {
			return err
		}

		var origRecord skydb.Record
		copyRecord(&origRecord, dbRecord)
		injectSigner(&origRecord, req.AssetStore)
		originalRecordMap[origRecord.ID] = &origRecord

		mergeRecord(dbRecord, record)
		*record = *dbRecord

		return
	})

	// execute before save hooks
	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			originalRecord, ok := originalRecordMap[record.ID]
			// FIXME: Hot-fix for issues #528
			// Defaults for record attributes should be provided
			// before executing hooks
			if !ok {
				record.OwnerID = req.UserInfo.ID
			}

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
		now := timeNow()

		var deltaRecord skydb.Record
		originalRecord, ok := originalRecordMap[record.ID]
		if !ok {
			originalRecord = &skydb.Record{}

			record.OwnerID = req.UserInfo.ID
			record.CreatedAt = now
			record.CreatorID = req.UserInfo.ID
		}

		record.UpdatedAt = now
		record.UpdaterID = req.UserInfo.ID

		deriveDeltaRecord(&deltaRecord, originalRecord, record)

		if dbErr := db.SaveDeltaRecord(&deltaRecord, originalRecord, record); dbErr != nil {
			err = skyerr.NewError(skyerr.UnexpectedError, dbErr.Error())
		}
		injectSigner(&deltaRecord, req.AssetStore)
		*record = deltaRecord

		return
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		return skyerr.NewError(skyerr.UnexpectedError, "atomic operation failed")
	}

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

func copyRecord(dst, src *skydb.Record) {
	*dst = *src

	dst.Data = map[string]interface{}{}
	for key, value := range src.Data {
		dst.Data[key] = value
	}
}

func mergeRecord(dst, src *skydb.Record) {
	dst.ID = src.ID
	dst.ACL = src.ACL

	if src.DatabaseID != "" {
		dst.DatabaseID = src.DatabaseID
	}

	if dst.Data == nil {
		dst.Data = map[string]interface{}{}
	}

	for key, value := range src.Data {
		dst.Data[key] = value
	}
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

	var records []*skydb.Record
	for _, recordID := range recordIDs {
		if recordID.Type == db.UserRecordType() {
			resp.ErrMap[recordID] = skyerr.NewError(skyerr.PermissionDenied, "cannot delete user record")
			continue
		}

		var record skydb.Record
		if dbErr := db.Get(recordID, &record); dbErr != nil {
			if dbErr == skydb.ErrRecordNotFound {
				resp.ErrMap[recordID] = skyerr.NewError(skyerr.ResourceNotFound, "record not found")
			} else {
				resp.ErrMap[recordID] = skyerr.NewError(skyerr.UnexpectedError, dbErr.Error())
			}
		} else {
			if req.WithMasterKey || record.Accessible(req.UserInfo, skydb.WriteLevel) {
				records = append(records, &record)
			} else {
				resp.ErrMap[recordID] = skyerr.NewError(
					skyerr.PermissionDenied,
					"no permission to delete",
				)
			}
		}
	}

	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			err = req.HookRegistry.ExecuteHooks(req.Context, hook.BeforeDelete, record, nil)
			return
		})
	}

	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {

		if dbErr := db.Delete(record.ID); dbErr != nil {
			return skyerr.NewError(skyerr.UnexpectedError, dbErr.Error())
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
		switch val := value.(type) {
		default:
			log.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
			}).Panicf("got unrecgonized type = %T", value)
		case nil:
			// do nothing
		case int64:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeInteger,
			}
		case float64:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeNumber,
			}
		case string:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeString,
			}
		case time.Time:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeDateTime,
			}
		case bool:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeBoolean,
			}
		case *skydb.Asset:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeAsset,
			}
		case skydb.Reference:
			v := value.(skydb.Reference)
			schema[key] = skydb.FieldType{
				Type:          skydb.TypeReference,
				ReferenceType: v.Type(),
			}
		case skydb.Location:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeLocation,
			}
		case skydb.Sequence:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeSequence,
			}
		case skydb.Geometry:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeGeometry,
			}
		case skydb.Unknown:
			schema[key] = skydb.FieldType{
				Type:           skydb.TypeUnknown,
				UnderlyingType: val.UnderlyingType,
			}
		case map[string]interface{}, []interface{}:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeJSON,
			}
		}
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
