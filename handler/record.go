package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/ourd/asset"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
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
	for key, value := range r.Data {
		switch v := value.(type) {
		case time.Time:
			m[key] = transportDate(v)
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
		return r.InitFromMap(m)
	}

	return fmt.Errorf("record: want a dictionary, got %T", i)
}

func (r *transportRecord) InitFromMap(m map[string]interface{}) error {
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
	data, err := walkData(m)
	if err != nil {
		return err
	}
	r.Data = data

	return nil
}

type transportDate time.Time

func (date transportDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string    `json:"$type"`
		Date time.Time `json:"$date"`
	}{"date", time.Time(date)})
}

type transportAsset oddb.Asset

func (asset transportAsset) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"$type"`
		Name string `json:"$name"`
	}{"asset", asset.Name})
}

func purgeReservedKey(m map[string]interface{}) {
	for key := range m {
		if key[0] == '_' {
			delete(m, key)
		}
	}
}

func walkData(m map[string]interface{}) (mapReturned map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	return walkMap(m), err
}

func walkMap(m map[string]interface{}) map[string]interface{} {
	for key, value := range m {
		m[key] = parseInterface(value)
	}

	return m
}

func walkSlice(items []interface{}) []interface{} {
	for i, item := range items {
		items[i] = parseInterface(item)
	}

	return items
}

func parseInterface(i interface{}) interface{} {
	switch value := i.(type) {
	default:
		// considered a bug if this line is reached
		panic(fmt.Errorf("unsupported value = %T", value))
	case nil, bool, float64, string:
		// the set of value that json unmarshaller returns
		// http://golang.org/pkg/encoding/json/#Unmarshal
		return value
	case map[string]interface{}:
		kindi, typed := value["$type"]
		if !typed {
			// regular dictionary, go deeper
			return walkMap(value)
		}

		kind, ok := kindi.(string)
		if !ok {
			panic(fmt.Errorf(`got "$type"'s type = %T, want string`, kindi))
		}

		switch kind {
		case "keypath":
			panic(fmt.Errorf("unsupported $type of persistence = %s", kind))
		case "geo", "blob":
			panic(fmt.Errorf("unimplemented $type = %s", kind))
		case "asset":
			return parseAsset(value)
		case "ref":
			return parseRef(value)
		case "date":
			return parseDate(value)
		default:
			panic(fmt.Errorf("unknown $type = %s", kind))
		}
	case []interface{}:
		return walkSlice(value)
	}
}

func parseKeyPath(i interface{}) (keypath string, err error) {
	switch value := i.(type) {
	case map[string]interface{}:
		kindi, typed := value["$type"]
		if typed {
			kind, ok := kindi.(string)
			if ok && kind == "keypath" {
				keypath, ok = value["$val"].(string)
				return
			}
		}
	}

	err = errors.New("not a keypath")
	return
}

func parseExpression(i interface{}) oddb.Expression {
	if keypath, err := parseKeyPath(i); err == nil {
		return oddb.Expression{
			Type:  oddb.KeyPath,
			Value: keypath,
		}
	}

	return oddb.Expression{
		Type:  oddb.Literal,
		Value: parseInterface(i),
	}
}

func parseDate(m map[string]interface{}) time.Time {
	datei, ok := m["$date"]
	if !ok {
		panic(errors.New("missing compulsory field $date"))
	}
	dateStr, ok := datei.(string)
	if !ok {
		panic(fmt.Errorf("got type($date) = %T, want string", datei))
	}
	dt, err := time.Parse(time.RFC3339Nano, dateStr)
	if err != nil {
		panic(fmt.Errorf("failed to parse $date = %#v", dateStr))
	}

	return dt.In(time.UTC)
}

func parseAsset(m map[string]interface{}) oddb.Asset {
	namei, ok := m["$name"]
	if !ok {
		panic(errors.New("missing compulsory field $name"))
	}
	name, ok := namei.(string)
	if !ok {
		panic(fmt.Errorf("got type($name) = %T, want string", namei))
	}
	if name == "" {
		panic(errors.New("asset's $name should not be empty"))
	}

	return oddb.Asset{
		Name: name,
	}
}

func parseRef(m map[string]interface{}) oddb.Reference {
	idi, ok := m["$id"]
	if !ok {
		panic(errors.New("referencing without $id"))
	}
	id, ok := idi.(string)
	if !ok {
		panic(fmt.Errorf("got reference type($id) = %T, want string", idi))
	}
	ss := strings.SplitN(id, "/", 2)
	if len(ss) == 1 {
		panic(fmt.Errorf(`ref: "_id" should be of format '{type}/{id}', got %#v`, id))
	}
	return oddb.NewReference(ss[0], ss[1])
}

type responseItem struct {
	id     string
	record serializedRecord
	err    oderr.Error
}

func newResponseItem(record serializedRecord) responseItem {
	return responseItem{
		id:     record.Record.ID.String(),
		record: record,
	}
}

func newResponseItemErr(id string, err oderr.Error) responseItem {
	return responseItem{
		id:  id,
		err: err,
	}
}

func (item responseItem) MarshalJSON() ([]byte, error) {
	var (
		buf bytes.Buffer
		i   interface{}
	)
	if item.id != "" {
		buf.Write([]byte(`{"_id":"`))
		buf.WriteString(item.id)
		buf.Write([]byte(`",`))
	} else {
		buf.WriteRune('{')
	}
	buf.Write([]byte(`"_type":"`))
	if item.err != nil {
		buf.Write([]byte(`error"`))
		i = item.err
	} else if item.record.Record != nil {
		buf.Write([]byte(`record"`))
		i = item.record
	} else {
		panic(errors.New("inconsistent state: both err and record is nil"))
	}

	bodyBytes, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	if len(bodyBytes) > 2 {
		if bodyBytes[0] != '{' {
			return nil, fmt.Errorf("first char of embedded json != {: %v", string(bodyBytes))
		} else if bodyBytes[len(bodyBytes)-1] != '}' {
			return nil, fmt.Errorf("last char of embedded json != }: %v", string(bodyBytes))
		}
		buf.WriteByte(',')
		buf.Write(bodyBytes[1:])
	} else {
		buf.WriteByte('}')
	}

	return buf.Bytes(), nil
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
	db := payload.Database
	recordMaps, ok := payload.Data["records"].([]interface{})
	if !ok {
		response.Err = oderr.NewRequestInvalidErr(errors.New("expected list of record"))
		return
	}

	// slice to keep the order of incoming record id / error during parsing
	incomingRecordItems := make([]interface{}, 0, len(recordMaps))

	// valid records that throughout the handler
	records := []*oddb.Record{}

	for _, recordMap := range recordMaps {
		var record oddb.Record
		if err := (*transportRecord)(&record).InitFromJSON(recordMap); err != nil {
			incomingRecordItems = append(incomingRecordItems, err)
		} else {
			incomingRecordItems = append(incomingRecordItems, record.ID)
			records = append(records, &record)
		}
	}

	// keep the error produced for a recordID throughout the handler
	recordIDErrMap := map[oddb.RecordID]error{}
	originalRecordMap := map[oddb.RecordID]*oddb.Record{}

	// fetch records
	records = executeRecordFunc(records, recordIDErrMap, func(record *oddb.Record) (err error) {
		record.OwnerID = payload.UserInfoID
		var dbRecord oddb.Record
		err = db.Get(record.ID, &dbRecord)
		if err == oddb.ErrRecordNotFound {
			originalRecordMap[record.ID] = &oddb.Record{}
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
	if payload.HookRegistry != nil {
		records = executeRecordFunc(records, recordIDErrMap, func(record *oddb.Record) (err error) {
			err = payload.HookRegistry.ExecuteHooks(hook.BeforeSave, record)
			return
		})
	}

	// derive and extend record schema
	if err := extendRecordSchema(db, records); err != nil {
		log.Debugln(err)
		response.Err = oderr.ErrDatabaseSchemaMigrationFailed
		return
	}

	// save records
	records = executeRecordFunc(records, recordIDErrMap, func(record *oddb.Record) (err error) {
		var deltaRecord oddb.Record
		originalRecord, ok := originalRecordMap[record.ID]
		if !ok {
			panic(fmt.Sprintf("original record not found; recordID = %s", record.ID))
		}
		deriveDeltaRecord(&deltaRecord, originalRecord, record)

		err = db.Save(&deltaRecord)
		return
	})

	// execute after save hooks
	if payload.HookRegistry != nil {
		records = executeRecordFunc(records, recordIDErrMap, func(record *oddb.Record) (err error) {
			payload.HookRegistry.ExecuteHooks(hook.AfterSave, record)
			return
		})
	}

	currRecordIdx := 0
	results := make([]responseItem, 0, len(incomingRecordItems))
	for _, itemi := range incomingRecordItems {
		var result responseItem

		switch item := itemi.(type) {
		case error:
			result = newResponseItemErr("", oderr.NewRequestInvalidErr(item))
		case oddb.RecordID:
			if err, ok := recordIDErrMap[item]; ok {
				log.WithFields(log.Fields{
					"recordID": item,
					"err":      err,
				}).Debugln("failed to save record")

				result = newResponseItemErr(item.String(), oderr.NewResourceSaveFailureErrWithStringID("record", item.String()))
			} else {
				record := records[currRecordIdx]
				currRecordIdx++
				result = newResponseItem(newSerializedRecord(record, payload.AssetStore))
			}
		default:
			panic(fmt.Sprintf("unknown type of incoming item: %T", itemi))
		}

		results = append(results, result)
	}

	response.Result = results
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

	results := make([]responseItem, length, length)
	for i, recordID := range recordIDs {
		record := oddb.Record{}
		if err := db.Get(recordID, &record); err != nil {
			if err == oddb.ErrRecordNotFound {
				results[i] = newResponseItemErr(
					recordID.String(),
					oderr.ErrRecordNotFound,
				)
			} else {
				log.WithFields(log.Fields{
					"recordID": recordID,
					"err":      err,
				}).Errorln("Failed to fetch record")
				results[i] = newResponseItemErr(
					recordID.String(),
					oderr.NewResourceFetchFailureErr("record", recordID.String()),
				)
			}
		} else {
			results[i] = newResponseItem(newSerializedRecord(&record, payload.AssetStore))
		}
	}

	response.Result = results
}

func keyPathFromRaw(rawKeyPath map[string]interface{}) string {
	mapType := rawKeyPath["$type"]
	if mapType != "keypath" {
		panic(fmt.Errorf("got key path's type %v, want \"keypath\"", mapType))
	}

	keypath := rawKeyPath["$val"].(string)
	if keypath == "" {
		panic(errors.New("empty key path value"))
	}

	return keypath
}

func sortFromRaw(rawSort []interface{}, sort *oddb.Sort) {
	keyPathMap, _ := rawSort[0].(map[string]interface{})
	if len(keyPathMap) == 0 {
		panic(errors.New("empty key path in sort descriptor"))
	}
	keyPath := keyPathFromRaw(keyPathMap)

	orderStr, _ := rawSort[1].(string)
	if orderStr == "" {
		panic(errors.New("empty sort order in sort descriptor"))
	}

	var sortOrder oddb.SortOrder
	switch orderStr {
	case "asc":
		sortOrder = oddb.Asc
	case "desc":
		sortOrder = oddb.Desc
	default:
		panic(fmt.Errorf("unknown sort order: %v", orderStr))
	}

	sort.KeyPath = keyPath
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

	eagerKeys := []string{}
	if keyPaths, ok := payload.Data["eager"].([]interface{}); ok {
		for _, keyPathRaw := range keyPaths {
			if keypath, err := parseKeyPath(keyPathRaw); err == nil {
				if strings.Contains(keypath, ".") {
					response.Err = oderr.NewRequestInvalidErr(errors.New("multi level eager loading not supported"))
					return
				}
				eagerKeys = append(eagerKeys, keypath)
			} else {
				response.Err = oderr.NewRequestInvalidErr(errors.New("invalid key path format"))
				return
			}
		}
	}

	results, err := db.Query(&query)
	if err != nil {
		response.Err = oderr.ErrDatabaseOpenFailed
		return
	}
	defer results.Close()

	records := []responseItem{}
	eagerLoadIDs := []oddb.RecordID{}
	for results.Scan() {
		record := results.Record()
		for _, eagerKey := range eagerKeys {
			if ref, ok := record.Data[eagerKey].(oddb.Reference); ok {
				eagerLoadIDs = append(eagerLoadIDs, ref.ID)
			}
		}
		records = append(records, newResponseItem(newSerializedRecord(&record, payload.AssetStore)))
	}

	if err != nil {
		response.Err = oderr.ErrDatabaseQueryFailed
		return
	}

	eagerRecords := []responseItem{}
	for _, rid := range eagerLoadIDs {
		record := oddb.Record{}
		if err := db.Get(rid, &record); err == nil {
			eagerRecords = append(eagerRecords, newResponseItem(newSerializedRecord(&record, payload.AssetStore)))
		} else {
			log.WithFields(log.Fields{
				"ID":  rid,
				"err": err,
			}).Debugln("Unable to eager load record.")
		}
	}

	response.Result = records
	if len(eagerKeys) > 0 {
		response.OtherResult = map[string]interface{}{
			"eager_load": eagerRecords,
		}
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
func RecordDeleteHandler(payload *router.Payload, response *router.Response) {
	db := payload.Database

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

	var deleteFunc func(oddb.RecordID, *oddb.Record) error
	if payload.HookRegistry != nil {
		deleteFunc = func(recordID oddb.RecordID, record *oddb.Record) error {
			payload.HookRegistry.ExecuteHooks(hook.BeforeDelete, record)
			err := db.Delete(recordID)
			if err == nil {
				payload.HookRegistry.ExecuteHooks(hook.AfterDelete, record)
			}
			return err
		}
	} else {
		deleteFunc = func(recordID oddb.RecordID, record *oddb.Record) error {
			return db.Delete(recordID)
		}
	}

	results := make([]interface{}, 0, length)
	for _, recordID := range recordIDs {
		var (
			err    error
			item   interface{}
			record oddb.Record
		)

		err = db.Get(recordID, &record)

		if err == nil {
			err = deleteFunc(recordID, &record)
		}

		if err == nil {
			item = struct {
				ID   oddb.RecordID `json:"_id"`
				Type string        `json:"_type"`
			}{recordID, "record"}
		} else if err == oddb.ErrRecordNotFound {
			item = newResponseItemErr(
				recordID.String(),
				oderr.ErrRecordNotFound,
			)
		} else {
			item = newResponseItemErr(
				recordID.String(),
				oderr.NewResourceDeleteFailureErrWithStringID("record", recordID.String()),
			)
		}

		results = append(results, item)
	}

	response.Result = results
}

type compRecordID struct {
	kind string
	id   string
}

func (rid compRecordID) ID() string {
	return rid.kind + "/" + rid.id
}
