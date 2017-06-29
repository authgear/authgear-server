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
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	pluginEvent "github.com/skygeario/skygear-server/pkg/server/plugin/event"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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

	// Records contains the successfully de-serialized record
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
	HookRegistry  *hook.Registry     `inject:"HookRegistry"`
	AssetStore    asset.Store        `inject:"AssetStore"`
	AccessModel   skydb.AccessModel  `inject:"AccessModel"`
	EventSender   pluginEvent.Sender `inject:"PluginEventSender"`
	Authenticator router.Processor   `preprocessor:"authenticator"`
	DBConn        router.Processor   `preprocessor:"dbconn"`
	InjectUser    router.Processor   `preprocessor:"inject_user"`
	InjectDB      router.Processor   `preprocessor:"inject_db"`
	RequireUser   router.Processor   `preprocessor:"require_user"`
	PluginReady   router.Processor   `preprocessor:"plugin_ready"`
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

	resultFilter, err := newRecordResultFilter(
		payload.DBConn,
		h.AssetStore,
		payload.AuthInfo,
		payload.HasMasterKey(),
	)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	log.Debugf("Working with accessModel %v", h.AccessModel)

	req := recordModifyRequest{
		Db:            payload.Database,
		Conn:          payload.DBConn,
		AssetStore:    h.AssetStore,
		HookRegistry:  h.HookRegistry,
		AuthInfo:      payload.AuthInfo,
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
				log.WithFields(logrus.Fields{
					"recordID": item,
					"err":      err,
				}).Debugln("failed to save record")

				result = newSerializedError(item.String(), err)
			} else {
				record := resp.SavedRecords[currRecordIdx]
				currRecordIdx++
				result = resultFilter.JSONResult(record)
			}
		default:
			panic(fmt.Sprintf("unknown type of incoming item: %T", itemi))
		}

		results = append(results, result)
	}
	response.Result = results

	if resp.SchemaUpdated && h.EventSender != nil {
		err := sendSchemaChangedEvent(h.EventSender, payload.Database)
		if err != nil {
			log.WithField("err", err).Warn("Fail to send schema changed event")
		}
	}
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
	PluginReady   router.Processor  `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RecordFetchHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.PluginReady,
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
	resultFilter, err := newRecordResultFilter(
		payload.DBConn,
		h.AssetStore,
		payload.AuthInfo,
		payload.HasMasterKey(),
	)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	fetcher := newRecordFetcher(db, payload.DBConn, payload.HasMasterKey())

	results := make([]interface{}, p.ItemLen(), p.ItemLen())
	for i, recordID := range p.RecordIDs {
		record, err := fetcher.fetchRecord(recordID, payload.AuthInfo, skydb.ReadLevel)
		if err != nil {
			results[i] = newSerializedError(
				recordID.String(),
				err,
			)
			continue
		}
		results[i] = resultFilter.JSONResult(record)
	}

	response.Result = results
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
		return err
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
	PluginReady   router.Processor  `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RecordQueryHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.PluginReady,
	}
}

func (h *RecordQueryHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RecordQueryHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &recordQueryPayload{}
	parser := QueryParser{UserID: payload.AuthInfoID}
	skyErr := p.Decode(payload.Data, &parser)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if payload.AuthInfo != nil {
		p.Query.ViewAsUser = payload.AuthInfo
	}

	if payload.HasMasterKey() {
		p.Query.BypassAccessControl = true
	}

	fieldACL := func() skydb.FieldACL {
		acl, err := payload.DBConn.GetRecordFieldAccess()
		if err != nil {
			panic(err)
		}
		return acl
	}()

	if !p.Query.BypassAccessControl {
		visitor := &queryAccessVisitor{
			FieldACL:   fieldACL,
			RecordType: p.Query.Type,
			AuthInfo:   p.Query.ViewAsUser,
			ExpressionACLChecker: ExpressionACLChecker{
				FieldACL:   fieldACL,
				RecordType: p.Query.Type,
				AuthInfo:   payload.AuthInfo,
				Database:   payload.Database,
			},
		}
		p.Query.Accept(visitor)
		if err := visitor.Error(); err != nil {
			response.Err = err
			return
		}
	}

	db := payload.Database

	results, err := db.Query(&p.Query)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	defer results.Close()

	records := []skydb.Record{}
	for results.Scan() {
		record := results.Record()
		records = append(records, record)
	}

	if results.Err() != nil {
		response.Err = skyerr.MakeError(results.Err())
		return
	}

	// Scan does not query assets,
	// it only replaces them with assets then only have name,
	// so we replace them with some complete assets.
	makeAssetsComplete(db, payload.DBConn, records)

	eagerRecords := doQueryEager(db, eagerIDs(db, records, p.Query))

	recordResultFilter, err := newRecordResultFilter(
		payload.DBConn,
		h.AssetStore,
		payload.AuthInfo,
		p.Query.BypassAccessControl,
	)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	resultFilter := queryResultFilter{
		Database:           db,
		Query:              p.Query,
		EagerRecords:       eagerRecords,
		recordResultFilter: recordResultFilter,
	}

	output := make([]interface{}, len(records))
	for i := range records {
		record := records[i]
		output[i] = resultFilter.JSONResult(&record)
	}

	response.Result = output

	resultInfo, err := queryResultInfo(db, &p.Query, results)
	if err != nil {
		response.Err = skyerr.MakeError(err)
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
	PluginReady   router.Processor  `preprocessor:"plugin_ready"`
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
		AuthInfo:          payload.AuthInfo,
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
			log.WithFields(logrus.Fields{
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
