package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/ourd/asset"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbconv"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type serializedRecord struct {
	Record     *oddb.Record
	AssetStore asset.Store
}

func newSerializedRecord(record *oddb.Record, assetStore asset.Store) serializedRecord {
	return serializedRecord{record, assetStore}
}

func (s serializedRecord) MarshalJSON() ([]byte, error) {
	r := s.Record

	m := map[string]interface{}{}
	m["_id"] = s.Record.ID.String()
	m["_type"] = "record"

	for key, value := range r.Data {
		switch v := value.(type) {
		case time.Time:
			m[key] = oddbconv.ToMap(oddbconv.MapTime(v))
		case oddb.Reference:
			m[key] = oddbconv.ToMap(oddbconv.MapReference(v))
		case *oddb.Location:
			m[key] = oddbconv.ToMap((*oddbconv.MapLocation)(v))
		case oddb.Asset:
			// TODO: refactor out this if. We know whether we are
			// injected an asset store at the start of handler
			var url string
			if signer, ok := s.AssetStore.(asset.URLSigner); ok {
				url = signer.SignedURL(v.Name, time.Now().Add(15*time.Minute))
			} else {
				url = ""
			}
			m[key] = struct {
				Type string `json:"$type"`
				Name string `json:"$name"`
				URL  string `json:"$url,omitempty"`
			}{"asset", v.Name, url}
		default:
			m[key] = v
		}
	}

	if r.OwnerID != "" {
		m["_ownerID"] = r.OwnerID
	}
	m["_access"] = r.ACL

	transient := map[string]interface{}{}
	for key, value := range r.Transient {
		switch v := value.(type) {
		case oddb.Record:
			transient[key] = newSerializedRecord(&v, s.AssetStore)
		default:
			transient[key] = v
		}
	}
	if len(transient) > 0 {
		m["_transient"] = transient
	}

	return json.Marshal(m)
}

// transportRecord override JSON serialization and deserialization of
// oddb.Record
type transportRecord oddb.Record

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
		acl := oddb.RecordACL{}
		if err := acl.InitFromJSON(aclData); err != nil {
			return fmt.Errorf(`record/json: %v`, err)
		}
		r.ACL = acl
	}

	purgeReservedKey(m)
	data := map[string]interface{}{}
	if err := (*oddbconv.MapData)(&data).FromMap(m); err != nil {
		return err
	}
	r.Data = data

	return nil
}

func purgeReservedKey(m map[string]interface{}) {
	for key := range m {
		if key[0] == '_' {
			delete(m, key)
		}
	}
}

type jsonData map[string]interface{}

func (data jsonData) ToMap(m map[string]interface{}) {
	for key, value := range data {
		if mapper, ok := value.(oddbconv.ToMapper); ok {
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
	err oderr.Error
}

func newSerializedError(id string, err oderr.Error) serializedError {
	return serializedError{
		id:  id,
		err: err,
	}
}

func (s serializedError) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"_type":   "error",
		"type":    s.err.Type(),
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
func RecordSaveHandler(payload *router.Payload, response *router.Response) {
	var (
		records []*oddb.Record
		atomic  bool
	)
	atomic, _ = payload.Data["atomic"].(bool)

	recordMaps, ok := payload.Data["records"].([]interface{})
	if !ok {
		response.Err = oderr.NewRequestInvalidErr(errors.New("expected list of record"))
		return
	}

	// slice to keep the order of incoming record id / error during parsing
	incomingRecordItems := make([]interface{}, 0, len(recordMaps))

	for _, recordMap := range recordMaps {
		var record oddb.Record
		if err := (*transportRecord)(&record).InitFromJSON(recordMap); err != nil {
			incomingRecordItems = append(incomingRecordItems, err)
		} else {
			incomingRecordItems = append(incomingRecordItems, record.ID)
			records = append(records, &record)
		}
	}

	req := recordModifyRequest{
		Db:            payload.Database,
		HookRegistry:  payload.HookRegistry,
		UserInfoID:    payload.UserInfoID,
		RecordsToSave: records,
		Atomic:        atomic,
	}
	resp := recordModifyResponse{
		ErrMap: map[oddb.RecordID]error{},
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
			result = newSerializedError("", oderr.NewRequestInvalidErr(item))
		case oddb.RecordID:
			if err, ok := resp.ErrMap[item]; ok {
				log.WithFields(log.Fields{
					"recordID": item,
					"err":      err,
				}).Debugln("failed to save record")

				result = newSerializedError(item.String(), oderr.NewResourceSaveFailureErrWithStringID("record", item.String()))
			} else {
				record := resp.SavedRecords[currRecordIdx]
				currRecordIdx++
				result = newSerializedRecord(record, payload.AssetStore)
			}
		default:
			panic(fmt.Sprintf("unknown type of incoming item: %T", itemi))
		}

		results = append(results, result)
	}

	response.Result = results
}

type recordModifyFunc func(*recordModifyRequest, *recordModifyResponse) error

func atomicModifyFunc(req *recordModifyRequest, resp *recordModifyResponse, mFunc recordModifyFunc) recordModifyFunc {
	return func(req *recordModifyRequest, resp *recordModifyResponse) (err error) {
		txDB, ok := req.Db.(oddb.TxDatabase)
		if !ok {
			err = oderr.ErrDatabaseTxNotSupported
			return
		}

		err = withTransaction(txDB, func() error {
			return mFunc(req, resp)
		})

		if len(resp.ErrMap) > 0 {
			err = oderr.NewAtomicOperationFailedErr(resp.ErrMap)
		} else if err != nil {
			err = oderr.NewAtomicOperationFailedErrWithCause(err)
		}
		return
	}
}

func withTransaction(txDB oddb.TxDatabase, do func() error) (err error) {
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
	Db           oddb.Database
	HookRegistry *hook.Registry
	Atomic       bool

	// Save only
	RecordsToSave []*oddb.Record
	UserInfoID    string

	// Delete Only
	RecordIDsToDelete []oddb.RecordID
}

type recordModifyResponse struct {
	ErrMap           map[oddb.RecordID]error
	SavedRecords     []*oddb.Record
	DeletedRecordIDs []oddb.RecordID
}

func recordSaveHandler(req *recordModifyRequest, resp *recordModifyResponse) error {
	db := req.Db
	records := req.RecordsToSave

	// fetch records
	originalRecordMap := map[oddb.RecordID]*oddb.Record{}
	records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
		record.OwnerID = req.UserInfoID
		var dbRecord oddb.Record
		err = db.Get(record.ID, &dbRecord)
		if err == oddb.ErrRecordNotFound {
			return nil
		}

		var origRecord oddb.Record
		copyRecord(&origRecord, &dbRecord)
		originalRecordMap[origRecord.ID] = &origRecord

		mergeRecord(&dbRecord, record)
		*record = dbRecord
		return
	})

	// execute before save hooks
	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
			originalRecord, _ := originalRecordMap[record.ID]
			err = req.HookRegistry.ExecuteHooks(hook.BeforeSave, record, originalRecord)
			return
		})
	}

	// derive and extend record schema
	if err := extendRecordSchema(db, records); err != nil {
		return oderr.ErrDatabaseSchemaMigrationFailed
	}

	// save records
	records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
		var deltaRecord oddb.Record
		originalRecord, ok := originalRecordMap[record.ID]
		if !ok {
			originalRecord = &oddb.Record{}
		}
		deriveDeltaRecord(&deltaRecord, originalRecord, record)

		err = db.Save(&deltaRecord)
		return
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		return errors.New("atomic operation failed")
	}

	// execute after save hooks
	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
			originalRecord, _ := originalRecordMap[record.ID]
			req.HookRegistry.ExecuteHooks(hook.AfterSave, record, originalRecord)
			return
		})
	}

	resp.SavedRecords = records
	return nil
}

type recordFunc func(*oddb.Record) error

func executeRecordFunc(recordsIn []*oddb.Record, errMap map[oddb.RecordID]error, rFunc recordFunc) (recordsOut []*oddb.Record) {
	for _, record := range recordsIn {
		if err := rFunc(record); err != nil {
			errMap[record.ID] = err
		} else {
			recordsOut = append(recordsOut, record)
		}
	}

	return
}

func copyRecord(dst, src *oddb.Record) {
	*dst = *src

	dst.Data = map[string]interface{}{}
	for key, value := range src.Data {
		dst.Data[key] = value
	}
}

func mergeRecord(dst, src *oddb.Record) {
	dst.ID = src.ID
	dst.ACL = src.ACL

	if src.DatabaseID != "" {
		dst.DatabaseID = src.DatabaseID
	}
	if src.OwnerID != "" {
		dst.OwnerID = src.OwnerID
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
func deriveDeltaRecord(dst, base, delta *oddb.Record) {
	dst.ID = delta.ID
	dst.ACL = delta.ACL
	dst.OwnerID = delta.OwnerID

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

func extendRecordSchema(db oddb.Database, records []*oddb.Record) error {
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
	finalSchema oddb.RecordSchema
	err         error
}

func newSchemaMerger() schemaMerger {
	return schemaMerger{finalSchema: oddb.RecordSchema{}}
}

func (m *schemaMerger) Extend(schema oddb.RecordSchema) {
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

func (m schemaMerger) Schema() (oddb.RecordSchema, error) {
	return m.finalSchema, m.err
}

func deriveRecordSchema(m oddb.Data) oddb.RecordSchema {
	schema := oddb.RecordSchema{}
	for key, value := range m {
		switch value.(type) {
		default:
			log.WithFields(log.Fields{
				"key":   key,
				"value": value,
			}).Panicf("got unrecgonized type = %T", value)
		case float64:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeNumber,
			}
		case string:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeString,
			}
		case time.Time:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeDateTime,
			}
		case bool:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeBoolean,
			}
		case oddb.Asset:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeAsset,
			}
		case oddb.Reference:
			v := value.(oddb.Reference)
			schema[key] = oddb.FieldType{
				Type:          oddb.TypeReference,
				ReferenceType: v.Type(),
			}
		case *oddb.Location:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeLocation,
			}
		case map[string]interface{}, []interface{}:
			schema[key] = oddb.FieldType{
				Type: oddb.TypeJSON,
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
func RecordFetchHandler(payload *router.Payload, response *router.Response) {
	interfaces, ok := payload.Data["ids"].([]interface{})
	if !ok {
		response.Err = oderr.NewRequestInvalidErr(errors.New("expected list of id"))
		return
	}

	length := len(interfaces)
	recordIDs := make([]oddb.RecordID, length, length)
	for i, it := range interfaces {
		rawID, ok := it.(string)
		if !ok {
			response.Err = oderr.NewRequestInvalidErr(errors.New("expected string id"))
			return
		}

		ss := strings.SplitN(rawID, "/", 2)
		if len(ss) == 1 {
			response.Err = oderr.NewRequestInvalidErr(fmt.Errorf("invalid id format: %v", rawID))
			return
		}

		recordIDs[i].Type = ss[0]
		recordIDs[i].Key = ss[1]
	}

	db := payload.Database

	results := make([]interface{}, length, length)
	for i, recordID := range recordIDs {
		record := oddb.Record{}
		if err := db.Get(recordID, &record); err != nil {
			if err == oddb.ErrRecordNotFound {
				results[i] = newSerializedError(
					recordID.String(),
					oderr.ErrRecordNotFound,
				)
			} else {
				log.WithFields(log.Fields{
					"recordID": recordID,
					"err":      err,
				}).Errorln("Failed to fetch record")
				results[i] = newSerializedError(
					recordID.String(),
					oderr.NewResourceFetchFailureErr("record", recordID.String()),
				)
			}
		} else {
			results[i] = newSerializedRecord(&record, payload.AssetStore)
		}
	}

	response.Result = results
}

func sortFromRaw(rawSort []interface{}, sort *oddb.Sort) {
	var (
		keyPath   string
		funcExpr  oddb.Func
		sortOrder oddb.SortOrder
	)
	switch v := rawSort[0].(type) {
	case map[string]interface{}:
		if err := (*oddbconv.MapKeyPath)(&keyPath).FromMap(v); err != nil {
			panic(err)
		}
	case []interface{}:
		var err error
		funcExpr, err = parseFunc(v)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unexpected type of sort expression = %T", rawSort[0]))
	}

	orderStr, _ := rawSort[1].(string)
	if orderStr == "" {
		panic(errors.New("empty sort order in sort descriptor"))
	}
	switch orderStr {
	case "asc":
		sortOrder = oddb.Asc
	case "desc":
		sortOrder = oddb.Desc
	default:
		panic(fmt.Errorf("unknown sort order: %v", orderStr))
	}

	sort.KeyPath = keyPath
	sort.Func = funcExpr
	sort.Order = sortOrder
}

func sortsFromRaw(rawSorts []interface{}) []oddb.Sort {
	length := len(rawSorts)
	sorts := make([]oddb.Sort, length, length)

	for i := range rawSorts {
		sortSlice, _ := rawSorts[i].([]interface{})
		if len(sortSlice) != 2 {
			panic(fmt.Errorf("got len(sort descriptor) = %v, want 2", len(sortSlice)))
		}
		sortFromRaw(sortSlice, &sorts[i])
	}

	return sorts
}

func predicateOperatorFromString(operatorString string) oddb.Operator {
	switch operatorString {
	case "and":
		return oddb.And
	case "or":
		return oddb.Or
	case "not":
		return oddb.Not
	case "eq":
		return oddb.Equal
	case "gt":
		return oddb.GreaterThan
	case "lt":
		return oddb.LessThan
	case "gte":
		return oddb.GreaterThanOrEqual
	case "lte":
		return oddb.LessThanOrEqual
	case "neq":
		return oddb.NotEqual
	case "like":
		return oddb.Like
	default:
		panic(fmt.Errorf("unrecognized operator = %s", operatorString))
	}
}

func predicateFromRaw(rawPredicate []interface{}) oddb.Predicate {
	if len(rawPredicate) < 2 {
		panic(fmt.Errorf("got len(predicate) = %v, want at least 2", len(rawPredicate)))
	}

	rawOperator, ok := rawPredicate[0].(string)
	if !ok {
		panic(fmt.Errorf("got predicate[0]'s type = %T, want string", rawPredicate[0]))
	}

	operator := predicateOperatorFromString(rawOperator)
	children := make([]interface{}, len(rawPredicate)-1)
	for i := 1; i < len(rawPredicate); i++ {
		if operator.IsCompound() {
			subRawPredicate, ok := rawPredicate[i].([]interface{})
			if !ok {
				panic(fmt.Errorf("got non-dict in subpredicate at %v", i-1))
			}
			children[i-1] = predicateFromRaw(subRawPredicate)
		} else {
			expr := parseExpression(rawPredicate[i])
			if expr.Type == oddb.KeyPath && strings.Contains(expr.Value.(string), ".") {

				panic(fmt.Errorf("Key path `%s` is not supported.", expr.Value))
			}
			children[i-1] = expr
		}
	}

	if operator.IsBinary() && len(children) != 2 {
		panic(fmt.Errorf("Expected number of expressions be 2, got %v", len(children)))
	}

	predicate := oddb.Predicate{
		Operator: operator,
		Children: children,
	}
	return predicate
}

func parseExpression(i interface{}) oddb.Expression {
	switch v := i.(type) {
	case map[string]interface{}:
		var keyPath string
		if err := oddbconv.MapFrom(i, (*oddbconv.MapKeyPath)(&keyPath)); err == nil {
			return oddb.Expression{
				Type:  oddb.KeyPath,
				Value: keyPath,
			}
		}
	case []interface{}:
		if f, err := parseFunc(v); err == nil {
			return oddb.Expression{
				Type:  oddb.Function,
				Value: f,
			}
		}
	}

	return oddb.Expression{
		Type:  oddb.Literal,
		Value: oddbconv.ParseInterface(i),
	}
}

func parseFunc(s []interface{}) (f oddb.Func, err error) {
	keyword, _ := s[0].(string)
	if keyword != "func" {
		return nil, errors.New("not a function")
	}

	funcName, _ := s[1].(string)
	switch funcName {
	case "distance":
		f, err = parseDistanceFunc(s[2:])
	case "":
		return nil, errors.New("empty function name")
	default:
		return nil, fmt.Errorf("got unrecgonized function name = %s", funcName)
	}

	return
}

func parseDistanceFunc(s []interface{}) (*oddb.DistanceFunc, error) {
	if len(s) != 2 {
		return nil, fmt.Errorf("want 2 arguments for distance func, got %d", len(s))
	}

	var field string
	if err := oddbconv.MapFrom(s[0], (*oddbconv.MapKeyPath)(&field)); err != nil {
		return nil, fmt.Errorf("invalid key path: %v", err)
	}

	var location oddb.Location
	if err := oddbconv.MapFrom(s[1], (*oddbconv.MapLocation)(&location)); err != nil {
		return nil, fmt.Errorf("invalid location: %v", err)
	}

	return &oddb.DistanceFunc{
		Field:    field,
		Location: &location,
	}, nil
}

func queryFromPayload(payload *router.Payload, query *oddb.Query) (err oderr.Error) {
	defer func() {
		// use panic to escape from inner error
		if r := recover(); r != nil {
			if queryErr, ok := r.(error); ok {
				log.WithField("payload", payload).Debugln("failed to construct query")
				err = oderr.NewFmt(oderr.RequestInvalidErr, "failed to construct query: %v", queryErr.Error())
			} else {
				log.WithField("recovered", r).Errorln("panic recovered while constructing query")
				err = oderr.New(oderr.RequestInvalidErr, "error occurred while constructing query")
			}
		}
	}()
	recordType, _ := payload.Data["record_type"].(string)
	if recordType == "" {
		return oderr.New(oderr.RequestInvalidErr, "recordType cannot be empty")
	}
	query.Type = recordType

	predicateRaw, ok := payload.Data["predicate"].([]interface{})
	if ok {
		predicate := predicateFromRaw(predicateRaw)
		query.Predicate = &predicate
	}

	if rawSorts, ok := payload.Data["sort"]; ok {
		if rawSorts, ok := rawSorts.([]interface{}); ok {
			query.Sorts = sortsFromRaw(rawSorts)
		} else {
			return oderr.New(oderr.RequestInvalidErr, "order has to be an array")
		}
	}

	if transientIncludes, ok := payload.Data["include"]; ok {
		if transientIncludes, ok := transientIncludes.(map[string]interface{}); ok {
			query.ComputedKeys = map[string]oddb.Expression{}
			for key, value := range transientIncludes {
				query.ComputedKeys[key] = parseExpression(value)
			}
		}
	}

	if desiredKeys, ok := payload.Data["desired_keys"]; ok {
		if desiredKeys, ok := desiredKeys.([]interface{}); ok {
			query.DesiredKeys = make([]string, len(desiredKeys))
			for i, key := range desiredKeys {
				key, ok := key.(string)
				if !ok {
					err = oderr.New(oderr.RequestInvalidErr, "unexpected value in desired_keys")
					return
				}
				query.DesiredKeys[i] = key
			}
		}
	}

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
func RecordQueryHandler(payload *router.Payload, response *router.Response) {
	db := payload.Database

	query := oddb.Query{}
	if err := queryFromPayload(payload, &query); err != nil {
		response.Err = err
		return
	}

	if payload.Data["database_id"] == "_public" {
		query.ReadableBy = payload.UserInfo.ID
	}

	var (
		eagerTransientKey string
		eagerKeyPath      string
	)

	// Handler supports a type of transient field that eager load
	// a referened record, which is not currently supported by oddb.
	// This type of expression is taken out of ComputedKeys and
	// the wanted records are added to the transient field later.
	for key, value := range query.ComputedKeys {
		if value.Type == oddb.KeyPath {
			if eagerTransientKey != "" {
				response.Err = oderr.NewRequestInvalidErr(errors.New("eager loading for multiple keys is not supported"))
				return
			}
			eagerTransientKey = key
			eagerKeyPath = value.Value.(string)
		}
	}
	delete(query.ComputedKeys, eagerTransientKey)

	results, err := db.Query(&query)
	if err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}
	defer results.Close()

	records := []oddb.Record{}
	for results.Scan() {
		record := results.Record()
		records = append(records, record)
	}

	if results.Err() != nil {
		response.Err = oderr.NewUnknownErr(results.Err())
		return
	}
	output := make([]interface{}, len(records))
	for i := range records {
		record := records[i]
		if eagerTransientKey != "" {
			record.Transient = map[string]interface{}{}
			if ref, ok := record.Data[eagerKeyPath].(oddb.Reference); ok {
				eagerRecord := oddb.Record{}
				if err := db.Get(ref.ID, &eagerRecord); err != nil {
					log.WithFields(log.Fields{
						"ID":  ref.ID,
						"err": err,
					}).Debugln("Unable to eager load record.")
				}
				record.Transient[eagerTransientKey] = eagerRecord
			}
		}

		output[i] = newSerializedRecord(&record, payload.AssetStore)
	}

	response.Result = output
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
func RecordDeleteHandler(payload *router.Payload, response *router.Response) {
	atomic, _ := payload.Data["atomic"].(bool)

	interfaces, ok := payload.Data["ids"].([]interface{})
	if !ok {
		response.Err = oderr.NewRequestInvalidErr(errors.New("expected list of id"))
		return
	}

	length := len(interfaces)
	recordIDs := make([]oddb.RecordID, length, length)
	for i, it := range interfaces {
		rawID, ok := it.(string)
		if !ok {
			response.Err = oderr.NewRequestInvalidErr(errors.New("expected string id"))
			return
		}

		ss := strings.SplitN(rawID, "/", 2)
		if len(ss) == 1 {
			response.Err = oderr.NewRequestInvalidErr(fmt.Errorf("invalid id format: %v", rawID))
			return
		}

		recordIDs[i].Type = ss[0]
		recordIDs[i].Key = ss[1]
	}

	req := recordModifyRequest{
		Db:                payload.Database,
		HookRegistry:      payload.HookRegistry,
		RecordIDsToDelete: recordIDs,
		Atomic:            atomic,
	}
	resp := recordModifyResponse{
		ErrMap: map[oddb.RecordID]error{},
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
			if err == oddb.ErrRecordNotFound {
				result = newSerializedError(
					recordID.String(),
					oderr.ErrRecordNotFound,
				)
			} else {
				log.WithFields(log.Fields{
					"recordID": recordID,
					"err":      err,
				}).Debugln("failed to delete record")

				result = newSerializedError(
					recordID.String(),
					oderr.NewResourceDeleteFailureErrWithStringID("record", recordID.String()),
				)
			}
		} else {
			result = struct {
				ID   oddb.RecordID `json:"_id"`
				Type string        `json:"_type"`
			}{recordID, "record"}
		}

		results = append(results, result)
	}

	response.Result = results
}

func recordDeleteHandler(req *recordModifyRequest, resp *recordModifyResponse) error {
	db := req.Db
	recordIDs := req.RecordIDsToDelete

	var records []*oddb.Record
	for _, recordID := range recordIDs {
		var record oddb.Record
		if err := db.Get(recordID, &record); err != nil {
			resp.ErrMap[recordID] = err
		} else {
			records = append(records, &record)
		}
	}

	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
			err = req.HookRegistry.ExecuteHooks(hook.BeforeDelete, record, nil)
			return
		})
	}

	records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
		return db.Delete(record.ID)
	})

	if req.Atomic && len(resp.ErrMap) > 0 {
		return errors.New("atomic operation failed")
	}

	if req.HookRegistry != nil {
		records = executeRecordFunc(records, resp.ErrMap, func(record *oddb.Record) (err error) {
			req.HookRegistry.ExecuteHooks(hook.AfterDelete, record, nil)
			return
		})
	}

	for _, record := range records {
		resp.DeletedRecordIDs = append(resp.DeletedRecordIDs, record.ID)
	}
	return nil
}

type compRecordID struct {
	kind string
	id   string
}

func (rid compRecordID) ID() string {
	return rid.kind + "/" + rid.id
}
