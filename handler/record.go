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

package handler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/context"

	"github.com/skygeario/skygear-server/asset"
	"github.com/skygeario/skygear-server/plugin/hook"
	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skydb"
	"github.com/skygeario/skygear-server/skydb/skyconv"
	"github.com/skygeario/skygear-server/skyerr"
)

type jsonData map[string]interface{}

func (data jsonData) ToMap(m map[string]interface{}) {
	for key, value := range data {
		if mapper, ok := value.(skyconv.ToMapper); ok {
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

// recordSavePayload decode and validate incoming mapstructure. It will store
// infroamtion regarding the payload after decode. Don't resue the struct for
// another payload.
type recordSavePayload struct {
	Atomic bool `mapstructure:"atomic"`

	// RawMaps stores the original incoming `records`.
	RawMaps []map[string]interface{} `mapstructure:"records"`

	// IncomigItems contains de-serialized recordID or de-serialization error,
	// the item is one-one corresponding to RawMaps.
	IncomingItems []interface{}

	// Records contains the sucessfully de-serialized record
	Records []*skydb.Record

	// Errs is the array of de-serialization errors
	Errs []skyerr.Error

	// Clean s true iff all incoming records are in proper format, design to
	// used with Atomic when handling the payload
	Clean bool
}

func (payload *recordSavePayload) purgeReservedKey(m map[string]interface{}) {
	for key := range m {
		if key == "" || key[0] == '_' {
			delete(m, key)
		}
	}
}

func (payload *recordSavePayload) ItemLen() int {
	return len(payload.RawMaps)
}

func (payload *recordSavePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *recordSavePayload) Validate() skyerr.Error {
	if payload.ItemLen() == 0 {
		return skyerr.NewInvalidArgument("expected list of record", []string{"records"})
	}

	payload.Clean = true
	payload.Errs = []skyerr.Error{}
	payload.IncomingItems = []interface{}{}
	payload.Records = []*skydb.Record{}
	for _, recordMap := range payload.RawMaps {
		var record skydb.Record
		if err := payload.InitRecord(recordMap, &record); err != nil {
			payload.Clean = false
			payload.Errs = append(payload.Errs, err)
			payload.IncomingItems = append(payload.IncomingItems, err)
		} else {
			payload.IncomingItems = append(payload.IncomingItems, record.ID)
			payload.Records = append(payload.Records, &record)
		}
	}

	return nil
}

// InitRecord is duplicated of skyconv.record FromMap FIXME
func (payload *recordSavePayload) InitRecord(m map[string]interface{}, r *skydb.Record) skyerr.Error {
	rawID, ok := m["_id"].(string)
	if !ok {
		return skyerr.NewInvalidArgument("missing required fields", []string{"id"})
	}

	ss := strings.SplitN(rawID, "/", 2)
	if len(ss) == 1 {
		return skyerr.NewInvalidArgument(
			`record: "_id" should be of format '{type}/{id}', got "`+rawID+`"`,
			[]string{"id"},
		)
	}

	recordType, id := ss[0], ss[1]

	r.ID.Key = id
	r.ID.Type = recordType

	aclData, ok := m["_access"]
	if ok && aclData != nil {
		aclSlice, ok := aclData.([]interface{})
		if !ok {
			return skyerr.NewInvalidArgument("_access must be an array", []string{"_access"})
		}
		acl := skydb.RecordACL{}
		for _, v := range aclSlice {
			ace := skydb.RecordACLEntry{}
			typed, ok := v.(map[string]interface{})
			if !ok {
				return skyerr.NewInvalidArgument("invalid _access entry", []string{"_access"})
			}
			if err := (*skyconv.MapACLEntry)(&ace).FromMap(typed); err != nil {
				return skyerr.NewInvalidArgument("invalid _access entry", []string{"_access"})
			}
			acl = append(acl, ace)
		}
		r.ACL = acl
	}

	payload.purgeReservedKey(m)
	data := map[string]interface{}{}
	if err := (*skyconv.MapData)(&data).FromMap(m); err != nil {
		return skyerr.NewError(skyerr.InvalidArgument, err.Error())
	}
	r.Data = data

	return nil
}

/*
RecordSaveHandler is dummy implementation on save/modify Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:save",
    "access_token": "validToken",
    "database_id": "_public",
    "records": [{
        "_id": "note/EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8",
        "content": "ewdsa",
        "_access": [{
            "role": "admin",
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
  "database_id": "_public",
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
      "_type": "record",
      "_access": [{
          "role": "admin",
          "level": "write"
      }]
    }
  ]
}
EOF
*/
type RecordSaveHandler struct {
	HookRegistry  *hook.Registry    `inject:"HookRegistry"`
	AssetStore    asset.Store       `inject:"AssetStore"`
	AccessModel   skydb.AccessModel `inject:"AccessModel"`
	Authenticator router.Processor  `preprocessor:"authenticator"`
	DBConn        router.Processor  `preprocessor:"dbconn"`
	InjectUser    router.Processor  `preprocessor:"inject_user"`
	InjectDB      router.Processor  `preprocessor:"inject_db"`
	RequireUser   router.Processor  `preprocessor:"require_user"`
	PluginReady   router.Processor  `preprocessor:"plugin"`
	preprocessors []router.Processor
}

func (h *RecordSaveHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
		h.PluginReady,
	}
}

func (h *RecordSaveHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RecordSaveHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &recordSavePayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if payload.Database.IsReadOnly() {
		response.Err = skyerr.NewError(skyerr.NotSupported, "modifying the selected database is not supported")
		return
	}

	log.Debugf("Working with accessModel %v", h.AccessModel)

	req := recordModifyRequest{
		Db:            payload.Database,
		Conn:          payload.DBConn,
		AssetStore:    h.AssetStore,
		HookRegistry:  h.HookRegistry,
		UserInfo:      payload.UserInfo,
		RecordsToSave: p.Records,
		Atomic:        p.Atomic,
		WithMasterKey: payload.HasMasterKey(),
		Context:       payload.Context,
	}
	resp := recordModifyResponse{
		ErrMap: map[skydb.RecordID]skyerr.Error{},
	}

	var saveFunc recordModifyFunc
	if p.Atomic {
		if !p.Clean {
			response.Err = skyerr.NewErrorWithInfo(
				skyerr.InvalidArgument,
				"fails to de-serialize records",
				map[string]interface{}{
					"arguments": "records",
					"errors":    p.Errs,
				})
			return
		}
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
	results := make([]interface{}, 0, p.ItemLen())
	for _, itemi := range p.IncomingItems {
		var result interface{}

		switch item := itemi.(type) {
		case skyerr.Error:
			result = newSerializedError("", item)
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
				result = (*skyconv.JSONRecord)(record)
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
	if f.db.Get(recordID, &dbRecord) == skydb.ErrRecordNotFound {
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

	record = &dbRecord
	if !f.withMasterKey && !dbRecord.Accessible(userInfo, skydb.WriteLevel) {
		err = skyerr.NewError(
			skyerr.PermissionDenied,
			"no permission to modify",
		)
	}

	return
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

			record.OwnerID = req.UserInfo.ID
			record.CreatedAt = now
			record.CreatorID = req.UserInfo.ID
		}

		record.UpdatedAt = now
		record.UpdaterID = req.UserInfo.ID

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
			err = req.HookRegistry.ExecuteHooks(req.Context, hook.AfterSave, record, originalRecord)
			if err != nil {
				log.Errorf("Error occurred while executing hooks: %s", err)
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

type recordFetchPayload struct {
	RecordIDs []skydb.RecordID
	RawIDs    []string `mapstructure:"ids"`
}

func (payload *recordFetchPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *recordFetchPayload) Validate() skyerr.Error {
	if len(payload.RawIDs) == 0 {
		return skyerr.NewInvalidArgument("expected list of id", []string{"ids"})
	}

	length := len(payload.RawIDs)
	payload.RecordIDs = make([]skydb.RecordID, length, length)
	for i, rawID := range payload.RawIDs {
		ss := strings.SplitN(rawID, "/", 2)
		if len(ss) == 1 {
			return skyerr.NewInvalidArgument(fmt.Sprintf("invalid id format: %v", rawID), []string{"ids"})
		}

		payload.RecordIDs[i].Type = ss[0]
		payload.RecordIDs[i].Key = ss[1]
	}
	return nil
}

func (payload *recordFetchPayload) ItemLen() int {
	return len(payload.RecordIDs)
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
	AssetStore    asset.Store       `inject:"AssetStore"`
	AccessModel   skydb.AccessModel `inject:"AccessModel"`
	Authenticator router.Processor  `preprocessor:"authenticator"`
	DBConn        router.Processor  `preprocessor:"dbconn"`
	InjectUser    router.Processor  `preprocessor:"inject_user"`
	InjectDB      router.Processor  `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *RecordFetchHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
	}
}

func (h *RecordFetchHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RecordFetchHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &recordFetchPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := payload.Database

	results := make([]interface{}, p.ItemLen(), p.ItemLen())
	for i, recordID := range p.RecordIDs {
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
			if payload.HasMasterKey() || record.Accessible(payload.UserInfo, skydb.ReadLevel) {
				injectSigner(&record, h.AssetStore)
				results[i] = (*skyconv.JSONRecord)(&record)
			} else {
				results[i] = newSerializedError(
					recordID.String(),
					skyerr.NewError(skyerr.PermissionDenied, "no permission to read"),
				)
			}
		}
	}

	response.Result = results
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

type recordQueryPayload struct {
	Query skydb.Query
}

func (payload *recordQueryPayload) Decode(data map[string]interface{}, parser *QueryParser) skyerr.Error {
	// Since the fields of skydb.Query is specified in the top-level,
	// we parse the data without mapstructure.
	// mapstructure "squash" tag does not work because skydb.Query
	// can only be converted using a hook func.

	if err := parser.queryFromRaw(data, &payload.Query); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload.Validate()
}

func (payload *recordQueryPayload) Validate() skyerr.Error {
	return nil
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
	AssetStore    asset.Store       `inject:"AssetStore"`
	AccessModel   skydb.AccessModel `inject:"AccessModel"`
	Authenticator router.Processor  `preprocessor:"authenticator"`
	DBConn        router.Processor  `preprocessor:"dbconn"`
	InjectUser    router.Processor  `preprocessor:"inject_user"`
	InjectDB      router.Processor  `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *RecordQueryHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
	}
}

func (h *RecordQueryHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RecordQueryHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &recordQueryPayload{}
	parser := QueryParser{UserID: payload.UserInfoID}
	skyErr := p.Decode(payload.Data, &parser)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if payload.UserInfo != nil {
		p.Query.ViewAsUser = payload.UserInfo
	}

	if payload.HasMasterKey() {
		p.Query.BypassAccessControl = true
	}

	db := payload.Database

	results, err := db.Query(&p.Query)
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

	eagers := eagerIDs(db, records, p.Query)
	eagerRecords := doQueryEager(db, eagers)

	output := make([]interface{}, len(records))
	for i := range records {
		record := records[i]

		for transientKey, transientExpression := range p.Query.ComputedKeys {
			if transientExpression.Type != skydb.KeyPath {
				continue
			}

			keyPath := transientExpression.Value.(string)
			val := record.Get(keyPath)
			var transientValue interface{}
			if val != nil {
				id := eagers[keyPath][i]
				eagerRecord := eagerRecords[keyPath][id.Key]
				if eagerRecord != nil {
					injectSigner(eagerRecord, h.AssetStore)
					transientValue = (*skyconv.JSONRecord)(eagerRecord)
				}
			}

			if record.Transient == nil {
				record.Transient = map[string]interface{}{}
			}
			record.Transient[transientKey] = transientValue
		}

		injectSigner(&record, h.AssetStore)
		output[i] = (*skyconv.JSONRecord)(&record)
	}

	response.Result = output

	resultInfo, err := queryResultInfo(db, &p.Query, results)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	if len(resultInfo) > 0 {
		response.Info = resultInfo
	}
}

type recordDeletePayload struct {
	RawIDs    []string `mapstructure:"ids"`
	Atomic    bool     `mapstructure:"atomic"`
	RecordIDs []skydb.RecordID
}

func (payload *recordDeletePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *recordDeletePayload) Validate() skyerr.Error {
	if len(payload.RawIDs) == 0 {
		return skyerr.NewInvalidArgument("expected list of id", []string{"ids"})
	}

	length := payload.ItemLen()
	payload.RecordIDs = make([]skydb.RecordID, length, length)
	for i, rawID := range payload.RawIDs {
		ss := strings.SplitN(rawID, "/", 2)
		if len(ss) == 1 {
			return skyerr.NewInvalidArgument(
				`record: "_id" should be of format '{type}/{id}', got "`+rawID+`"`,
				[]string{"ids"},
			)
		}

		payload.RecordIDs[i].Type = ss[0]
		payload.RecordIDs[i].Key = ss[1]
	}
	return nil
}

func (payload *recordDeletePayload) ItemLen() int {
	return len(payload.RawIDs)
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
	HookRegistry  *hook.Registry    `inject:"HookRegistry"`
	AccessModel   skydb.AccessModel `inject:"AccessModel"`
	Authenticator router.Processor  `preprocessor:"authenticator"`
	DBConn        router.Processor  `preprocessor:"dbconn"`
	InjectUser    router.Processor  `preprocessor:"inject_user"`
	InjectDB      router.Processor  `preprocessor:"inject_db"`
	RequireUser   router.Processor  `preprocessor:"require_user"`
	PluginReady   router.Processor  `preprocessor:"plugin"`
	preprocessors []router.Processor
}

func (h *RecordDeleteHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
		h.PluginReady,
	}
}

func (h *RecordDeleteHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RecordDeleteHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &recordDeletePayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if payload.Database.IsReadOnly() {
		response.Err = skyerr.NewError(skyerr.NotSupported, "modifying the selected database is not supported")
		return
	}

	req := recordModifyRequest{
		Db:                payload.Database,
		Conn:              payload.DBConn,
		HookRegistry:      h.HookRegistry,
		RecordIDsToDelete: p.RecordIDs,
		Atomic:            p.Atomic,
		WithMasterKey:     payload.HasMasterKey(),
		Context:           payload.Context,
		UserInfo:          payload.UserInfo,
	}
	resp := recordModifyResponse{
		ErrMap: map[skydb.RecordID]skyerr.Error{},
	}

	var deleteFunc recordModifyFunc
	if p.Atomic {
		deleteFunc = atomicModifyFunc(&req, &resp, recordDeleteHandler)
	} else {
		deleteFunc = recordDeleteHandler
	}

	if err := deleteFunc(&req, &resp); err != nil {
		log.Debugf("Failed to delete records: %v", err)
		response.Err = err
		return
	}

	results := make([]interface{}, 0, p.ItemLen())
	for _, recordID := range p.RecordIDs {
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
