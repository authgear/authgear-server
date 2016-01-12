package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skydbconv"
	"github.com/oursky/skygear/skyerr"
	"golang.org/x/net/context"
)

type serializedRecord struct {
	Record *skydb.Record
}

func newSerializedRecord(record *skydb.Record) serializedRecord {
	return serializedRecord{record}
}

func (s serializedRecord) MarshalJSON() ([]byte, error) {
	r := s.Record

	m := map[string]interface{}{}
	m["_id"] = s.Record.ID.String()
	m["_type"] = "record"
	m["_access"] = r.ACL

	if r.OwnerID != "" {
		m["_ownerID"] = r.OwnerID
	}
	if !r.CreatedAt.IsZero() {
		m["_created_at"] = r.CreatedAt
	}
	if r.CreatorID != "" {
		m["_created_by"] = r.CreatorID
	}
	if !r.UpdatedAt.IsZero() {
		m["_updated_at"] = r.UpdatedAt
	}
	if r.UpdaterID != "" {
		m["_updated_by"] = r.UpdaterID
	}

	for key, value := range r.Data {
		switch v := value.(type) {
		case time.Time:
			m[key] = skydbconv.ToMap(skydbconv.MapTime(v))
		case skydb.Reference:
			m[key] = skydbconv.ToMap(skydbconv.MapReference(v))
		case skydb.Location:
			m[key] = skydbconv.ToMap(skydbconv.MapLocation(v))
		case *skydb.Asset:
			m[key] = skydbconv.ToMap((*skydbconv.MapAsset)(v))
		default:
			m[key] = v
		}
	}

	transient := s.marshalTransient(r.Transient)
	if len(transient) > 0 {
		m["_transient"] = transient
	}

	return json.Marshal(m)
}

func (s serializedRecord) marshalTransient(transient map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for key, value := range transient {
		switch v := value.(type) {
		case skydb.Record:
			m[key] = newSerializedRecord(&v)
		default:
			m[key] = v
		}
	}
	return m
}

// transportRecord override JSON serialization and deserialization of
// skydb.Record
type transportRecord skydb.Record

func (r *transportRecord) UnmarshalJSON(data []byte) error {
	object := map[string]interface{}{}
	err := json.Unmarshal(data, &object)

	if err != nil {
		return err
	}

	return r.InitFromJSON(object)
}

func (r *transportRecord) InitFromJSON(i interface{}) error {
	if m, ok := i.(map[string]interface{}); ok {
		return r.FromMap(m)
	}

	return fmt.Errorf("record: want a dictionary, got %T", i)
}

func (r *transportRecord) FromMap(m map[string]interface{}) error {
	rawID, ok := m["_id"].(string)
	if !ok {
		return errors.New(`record: required field "_id" not found`)
	}

	ss := strings.SplitN(rawID, "/", 2)
	if len(ss) == 1 {
		return fmt.Errorf(`record: "_id" should be of format '{type}/{id}', got %#v`, rawID)
	}

	recordType, id := ss[0], ss[1]

	r.ID.Key = id
	r.ID.Type = recordType

	aclData, ok := m["_access"]
	if ok {
		acl := skydb.RecordACL{}
		if err := acl.InitFromJSON(aclData); err != nil {
			return fmt.Errorf(`record/json: %v`, err)
		}
		r.ACL = acl
	}

	purgeReservedKey(m)
	data := map[string]interface{}{}
	if err := (*skydbconv.MapData)(&data).FromMap(m); err != nil {
		return err
	}
	r.Data = data

	return nil
}

func purgeReservedKey(m map[string]interface{}) {
	for key := range m {
		if key == "" || key[0] == '_' {
			delete(m, key)
		}
	}
}

type jsonData map[string]interface{}

func (data jsonData) ToMap(m map[string]interface{}) {
	for key, value := range data {
		if mapper, ok := value.(skydbconv.ToMapper); ok {
			valueMap := map[string]interface{}{}
			mapper.ToMap(valueMap)
			m[key] = valueMap
		} else {
			m[key] = value
		}
	}
}

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
	}
	if s.err.Info() != nil {
		m["info"] = s.err.Info()
	}

	return json.Marshal(m)
}

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

/*
RecordSaveHandler is dummy implementation on save/modify Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:save",
    "access_token": "validToken",
    "database_id": "_private",
    "records": [{
        "_id": "note/EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8",
        "content": "ewdsa",
        "_access": [{
            "relation": "friend",
            "level": "write"
        }]
    }]
}
EOF

Save with reference
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
  "action": "record:save",
  "database_id": "_private",
  "access_token": "986bee3b-8dd9-45c2-b40c-8b6ef274cf12",
  "records": [
    {
      "collection": {
        "$type": "ref",
        "$id": "collection/10"
      },
      "noteOrder": 1,
      "content": "hi",
      "_id": "note/71BAE736-E9C5-43CB-ADD1-D8633B80CAFA",
      "_type": "record"
    }
  ]
}
EOF
*/
type RecordSaveHandler struct {
	HookRegistry *hook.Registry    `inject:"HookRegistry"`
	AssetStore   asset.Store       `inject:"AssetStore"`
	AccessModel  skydb.AccessModel `inject:"AccessModel"`
}

func (h *RecordSaveHandler) Setup() {
	return
}

func (h *RecordSaveHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *RecordSaveHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("Working with accessModel %v", h.AccessModel)
	var (
		records []*skydb.Record
		atomic  bool
	)
	atomic, _ = payload.Data["atomic"].(bool)

	recordMaps, ok := payload.Data["records"].([]interface{})
	if !ok {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "expected list of record")
		return
	}

	// slice to keep the order of incoming record id / error during parsing
	incomingRecordItems := make([]interface{}, 0, len(recordMaps))

	for _, recordMap := range recordMaps {
		var record skydb.Record
		if err := (*transportRecord)(&record).InitFromJSON(recordMap); err != nil {
			incomingRecordItems = append(incomingRecordItems, err)
		} else {
			incomingRecordItems = append(incomingRecordItems, record.ID)
			records = append(records, &record)
		}
	}

	req := recordModifyRequest{
		Db:            payload.Database,
		AssetStore:    h.AssetStore,
		HookRegistry:  h.HookRegistry,
		UserInfoID:    payload.UserInfoID,
		RecordsToSave: records,
		Atomic:        atomic,
		Context:       payload.Context,
	}
	resp := recordModifyResponse{
		ErrMap: map[skydb.RecordID]skyerr.Error{},
	}

	var saveFunc recordModifyFunc
	if atomic {
		saveFunc = atomicModifyFunc(&req, &resp, recordSaveHandler)
	} else {
		saveFunc = recordSaveHandler
	}

	if err := saveFunc(&req, &resp); err != nil {
		log.Debugf("Failed to save records: %v", err)

		response.Err = err
		return
	}

	currRecordIdx := 0
	results := make([]interface{}, 0, len(incomingRecordItems))
	for _, itemi := range incomingRecordItems {
		var result interface{}

		switch item := itemi.(type) {
		case error:
			result = newSerializedError("", skyerr.NewError(skyerr.InvalidArgument, item.Error()))
		case skydb.RecordID:
			if err, ok := resp.ErrMap[item]; ok {
				log.WithFields(log.Fields{
					"recordID": item,
					"err":      err,
				}).Debugln("failed to save record")

				result = newSerializedError(item.String(), err)
			} else {
				record := resp.SavedRecords[currRecordIdx]
				currRecordIdx++
				result = newSerializedRecord(record)
			}
		default:
			panic(fmt.Sprintf("unknown type of incoming item: %T", itemi))
		}

		results = append(results, result)
	}

	response.Result = results
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
				info[recordID.String()] = err.Message()
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
	Db           skydb.Database
	AssetStore   asset.Store
	HookRegistry *hook.Registry
	Atomic       bool
	Context      context.Context

	// Save only
	RecordsToSave []*skydb.Record
	UserInfoID    string

	// Delete Only
	RecordIDsToDelete []skydb.RecordID
}

type recordModifyResponse struct {
	ErrMap           map[skydb.RecordID]skyerr.Error
	SavedRecords     []*skydb.Record
	DeletedRecordIDs []skydb.RecordID
}

func recordSaveHandler(req *recordModifyRequest, resp *recordModifyResponse) skyerr.Error {
	db := req.Db
	records := req.RecordsToSave

	// fetch records
	originalRecordMap := map[skydb.RecordID]*skydb.Record{}
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		var dbRecord skydb.Record
		dbErr := db.Get(record.ID, &dbRecord)
		if dbErr == skydb.ErrRecordNotFound {
			return nil
		}

		var origRecord skydb.Record
		copyRecord(&origRecord, &dbRecord)
		injectSigner(&origRecord, req.AssetStore)
		originalRecordMap[origRecord.ID] = &origRecord

		mergeRecord(&dbRecord, record)
		*record = dbRecord

		return
	})

	// execute before save hooks
	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			originalRecord, _ := originalRecordMap[record.ID]
			pluginErr := req.HookRegistry.ExecuteHooks(req.Context, hook.BeforeSave, record, originalRecord)
			if pluginErr != nil {
				err = skyerr.NewError(skyerr.UnexpectedError, pluginErr.Error())
			}
			return
		})
	}

	// derive and extend record schema
	if err := extendRecordSchema(db, records); err != nil {
		log.WithField("err", err).Errorln("failed to migrate record schema")
		return skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate record schema")
	}

	// remove bogus field, they are only for schema change
	for _, r := range records {
		for k, v := range r.Data {
			switch v.(type) {
			case skydb.Sequence:
				delete(r.Data, k)
			}
		}
	}

	// save records
	records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
		now := timeNow()

		var deltaRecord skydb.Record
		originalRecord, ok := originalRecordMap[record.ID]
		if !ok {
			originalRecord = &skydb.Record{}

			record.OwnerID = req.UserInfoID
			record.CreatedAt = now
			record.CreatorID = req.UserInfoID
		}

		record.UpdatedAt = now
		record.UpdaterID = req.UserInfoID

		deriveDeltaRecord(&deltaRecord, originalRecord, record)

		if dbErr := db.Save(&deltaRecord); dbErr != nil {
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
			pluginErr := req.HookRegistry.ExecuteHooks(req.Context, hook.AfterSave, record, originalRecord)
			if pluginErr != nil {
				log.Errorf("Error occurred while executing hooks: %s", pluginErr.Error())
			}
			return
		})
	}

	resp.SavedRecords = records
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
	dst.ACL = delta.ACL
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

func extendRecordSchema(db skydb.Database, records []*skydb.Record) error {
	recordSchemaMergerMap := map[string]schemaMerger{}
	for _, record := range records {
		recordType := record.ID.Type
		merger, ok := recordSchemaMergerMap[recordType]
		if !ok {
			merger = newSchemaMerger()
			recordSchemaMergerMap[recordType] = merger
		}

		merger.Extend(deriveRecordSchema(record.Data))
	}

	for recordType, merger := range recordSchemaMergerMap {
		schema, err := merger.Schema()
		if err != nil {
			return err
		}

		if err = db.Extend(recordType, schema); err != nil {
			return err
		}
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
		switch value.(type) {
		default:
			log.WithFields(log.Fields{
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
		case map[string]interface{}, []interface{}:
			schema[key] = skydb.FieldType{
				Type: skydb.TypeJSON,
			}
		}
	}

	return schema
}

/*
RecordFetchHandler is dummy implementation on fetching Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:fetch",
    "access_token": "validToken",
    "database_id": "_private",
    "ids": ["note/1004", "note/1005"]
}
EOF
*/
type RecordFetchHandler struct {
	AssetStore  asset.Store       `inject:"AssetStore"`
	AccessModel skydb.AccessModel `inject:"AccessModel"`
}

func (h *RecordFetchHandler) Setup() {
	return
}

func (h *RecordFetchHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *RecordFetchHandler) Handle(payload *router.Payload, response *router.Response) {
	interfaces, ok := payload.Data["ids"].([]interface{})
	if !ok {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "expected list of id")
		return
	}

	length := len(interfaces)
	recordIDs := make([]skydb.RecordID, length, length)
	for i, it := range interfaces {
		rawID, ok := it.(string)
		if !ok {
			response.Err = skyerr.NewError(skyerr.InvalidArgument, "expected string id")
			return
		}

		ss := strings.SplitN(rawID, "/", 2)
		if len(ss) == 1 {
			response.Err = skyerr.NewErrorf(skyerr.InvalidArgument, "invalid id format: %v", rawID)
			return
		}

		recordIDs[i].Type = ss[0]
		recordIDs[i].Key = ss[1]
	}

	db := payload.Database

	results := make([]interface{}, length, length)
	for i, recordID := range recordIDs {
		record := skydb.Record{}
		if err := db.Get(recordID, &record); err != nil {
			if err == skydb.ErrRecordNotFound {
				results[i] = newSerializedError(
					recordID.String(),
					skyerr.NewError(skyerr.ResourceNotFound, "record not found"),
				)
			} else {
				log.WithFields(log.Fields{
					"recordID": recordID,
					"err":      err,
				}).Errorln("Failed to fetch record")
				results[i] = newSerializedError(
					recordID.String(),
					skyerr.NewResourceFetchFailureErr("record", recordID.String()),
				)
			}
		} else {
			injectSigner(&record, h.AssetStore)
			results[i] = newSerializedRecord(&record)
		}
	}

	response.Result = results
}

func eagerIDs(records []skydb.Record, query skydb.Query) map[string][]skydb.RecordID {
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
			valueAtKeyPath, ok := record.Data[keyPath]
			if !ok || valueAtKeyPath == nil {
				continue
			}
			ref, ok := valueAtKeyPath.(skydb.Reference)
			if !ok {
				continue
			}
			eagers[keyPath][i] = ref.ID
		}
	}
	return eagers
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

/*
RecordQueryHandler is dummy implementation on fetching Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:query",
    "access_token": "validToken",
    "database_id": "_private",
    "record_type": "note",
    "sort": [
        [{"$val": "noteOrder", "$type": "desc"}, "asc"]
    ]
}
EOF
*/
type RecordQueryHandler struct {
	AssetStore  asset.Store       `inject:"AssetStore"`
	AccessModel skydb.AccessModel `inject:"AccessModel"`
}

func (h *RecordQueryHandler) Setup() {
	return
}

func (h *RecordQueryHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *RecordQueryHandler) Handle(payload *router.Payload, response *router.Response) {
	db := payload.Database

	query := skydb.Query{}
	parser := &QueryParser{
		UserID: payload.UserInfoID,
	}
	if err := parser.queryFromRaw(payload.Data, &query); err != nil {
		response.Err = err
		return
	}

	if payload.Data["database_id"] == "_public" {
		query.ReadableBy = payload.UserInfoID
	}

	results, err := db.Query(&query)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	defer results.Close()

	records := []skydb.Record{}
	for results.Scan() {
		record := results.Record()
		records = append(records, record)
	}

	if results.Err() != nil {
		response.Err = skyerr.NewUnknownErr(results.Err())
		return
	}

	eagers := eagerIDs(records, query)
	eagerRecords := doQueryEager(db, eagers)

	output := make([]interface{}, len(records))
	for i := range records {
		record := records[i]

		for transientKey, transientExpression := range query.ComputedKeys {
			if transientExpression.Type != skydb.KeyPath {
				continue
			}

			keyPath := transientExpression.Value.(string)
			if _, ok := record.Data[keyPath]; ok {
				if record.Transient == nil {
					record.Transient = map[string]interface{}{}
				}
				id := eagers[keyPath][i]
				eagerRecord := eagerRecords[keyPath][id.Key]
				if eagerRecord != nil {
					injectSigner(eagerRecord, h.AssetStore)
					record.Transient[transientKey] = newSerializedRecord(eagerRecord)
				} else {
					record.Transient[transientKey] = nil
				}
			}
		}

		injectSigner(&record, h.AssetStore)
		output[i] = newSerializedRecord(&record)
	}

	response.Result = output

	resultInfo, err := queryResultInfo(db, &query, results)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	if len(resultInfo) > 0 {
		response.Info = resultInfo
	}
}

/*
RecordDeleteHandler is dummy implementation on delete Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:delete",
    "access_token": "validToken",
    "database_id": "_private",
    "ids": ["note/EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8"]
}
EOF
*/
type RecordDeleteHandler struct {
	HookRegistry *hook.Registry    `inject:"HookRegistry"`
	AccessModel  skydb.AccessModel `inject:"AccessModel"`
}

func (h *RecordDeleteHandler) Setup() {
	return
}

func (h *RecordDeleteHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *RecordDeleteHandler) Handle(payload *router.Payload, response *router.Response) {
	atomic, _ := payload.Data["atomic"].(bool)

	interfaces, ok := payload.Data["ids"].([]interface{})
	if !ok {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "expected list of id")
		return
	}

	length := len(interfaces)
	recordIDs := make([]skydb.RecordID, length, length)
	for i, it := range interfaces {
		rawID, ok := it.(string)
		if !ok {
			response.Err = skyerr.NewError(skyerr.InvalidArgument, "expected string id")
			return
		}

		ss := strings.SplitN(rawID, "/", 2)
		if len(ss) == 1 {
			response.Err = skyerr.NewErrorf(skyerr.InvalidArgument, "invalid id format: %v", rawID)
			return
		}

		recordIDs[i].Type = ss[0]
		recordIDs[i].Key = ss[1]
	}

	req := recordModifyRequest{
		Db:                payload.Database,
		HookRegistry:      h.HookRegistry,
		RecordIDsToDelete: recordIDs,
		Atomic:            atomic,
		Context:           payload.Context,
	}
	resp := recordModifyResponse{
		ErrMap: map[skydb.RecordID]skyerr.Error{},
	}

	var deleteFunc recordModifyFunc
	if atomic {
		deleteFunc = atomicModifyFunc(&req, &resp, recordDeleteHandler)
	} else {
		deleteFunc = recordDeleteHandler
	}

	if err := deleteFunc(&req, &resp); err != nil {
		log.Debugf("Failed to delete records: %v", err)

		response.Err = err
		return
	}

	results := make([]interface{}, 0, length)
	for _, recordID := range recordIDs {
		var result interface{}

		if err, ok := resp.ErrMap[recordID]; ok {
			log.WithFields(log.Fields{
				"recordID": recordID,
				"err":      err,
			}).Debugln("failed to delete record")
			result = newSerializedError(
				recordID.String(),
				err,
			)
		} else {
			result = struct {
				ID   skydb.RecordID `json:"_id"`
				Type string         `json:"_type"`
			}{recordID, "record"}
		}

		results = append(results, result)
	}

	response.Result = results
}

func recordDeleteHandler(req *recordModifyRequest, resp *recordModifyResponse) skyerr.Error {
	db := req.Db
	recordIDs := req.RecordIDsToDelete

	var records []*skydb.Record
	for _, recordID := range recordIDs {
		var record skydb.Record
		if dbErr := db.Get(recordID, &record); dbErr != nil {
			if dbErr == skydb.ErrRecordNotFound {
				resp.ErrMap[recordID] = skyerr.NewError(skyerr.ResourceNotFound, "record not found")
			} else {
				resp.ErrMap[recordID] = skyerr.NewError(skyerr.UnexpectedError, dbErr.Error())
			}
		} else {
			records = append(records, &record)
		}
	}

	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *skydb.Record) (err skyerr.Error) {
			pluginErr := req.HookRegistry.ExecuteHooks(req.Context, hook.BeforeDelete, record, nil)
			if pluginErr != nil {
				err = skyerr.NewError(skyerr.UnexpectedError, pluginErr.Error())
			}
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
			pluginErr := req.HookRegistry.ExecuteHooks(req.Context, hook.AfterDelete, record, nil)
			if pluginErr != nil {
				log.Errorf("Error occurred while executing hooks: %s", pluginErr.Error())
			}
			return
		})
	}

	for _, record := range records {
		resp.DeletedRecordIDs = append(resp.DeletedRecordIDs, record.ID)
	}
	return nil
}
