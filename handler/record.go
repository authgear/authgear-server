package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"runtime"
	"strings"
	"time"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

// transportRecord override JSON serialization and deserialization of
// oddb.Record
type transportRecord oddb.Record

func (r transportRecord) MarshalJSON() ([]byte, error) {
	// NOTE(limouren): marshalling of type/key is delegated to responseItem
	if r.Data == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(transportData(r.Data))
}

type transportData map[string]interface{}

func (data transportData) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	for key, value := range data {
		switch v := value.(type) {
		case time.Time:
			m[key] = transportDate(v)
		default:
			m[key] = v
		}
	}
	return json.Marshal(m)
}

type transportDate time.Time

func (date transportDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string    `json:"$type"`
		Date time.Time `json:"$date"`
	}{"date", time.Time(date)})
}

func (r *transportRecord) UnmarshalJSON(data []byte) error {
	object := map[string]interface{}{}
	err := json.Unmarshal(data, &object)

	if err != nil {
		return err
	}

	return r.InitFromMap(object)
}

func (r *transportRecord) InitFromMap(m map[string]interface{}) error {
	rawID, ok := m["_id"].(string)
	if !ok {
		return errors.New(`record/json: required field "_id" not found`)
	}

	ss := strings.SplitN(rawID, "/", 2)
	if len(ss) == 1 {
		return fmt.Errorf(`record/json: "_id" should be of format '{type}/{id}', got %#v`, rawID)
	}

	recordType, id := ss[0], ss[1]

	r.ID.Key = id
	r.ID.Type = recordType

	purgeReservedKey(m)
	data, err := walkData(m)
	if err != nil {
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
	record *transportRecord
	err    oderr.Error
}

func newResponseItem(record *transportRecord) responseItem {
	return responseItem{
		id:     record.ID.String(),
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
	buf.Write([]byte(`{"_id":"`))
	buf.WriteString(item.id)
	buf.Write([]byte(`","_type":"`))
	if item.err != nil {
		buf.Write([]byte(`error"`))
		i = item.err
	} else if item.record != nil {
		buf.Write([]byte(`record"`))
		i = item.record
	} else {
		panic("inconsistent state: both err and record is nil")
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
  -d @- http://192.168.1.89/ <<EOF
{
    "action": "record:save",
    "access_token": "validToken",
    "database_id": "private",
    "records": [{
        "_id": "EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8",
        "_type": "note",
        "content": "ewdsa"
    }]
}
EOF

Save with reference
curl -X POST -H "Content-Type: application/json" \
  -d @- http://192.168.1.89/ <<EOF
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

	length := len(recordMaps)

	items := make([]recordSaveItem, 0, length)
	for _, recordMapI := range recordMaps {
		item := newRecordSaveItem(recordMapI)

		if err := (*transportRecord)(&item.record).InitFromMap(item.m); err != nil {
			item.err = oderr.NewRequestInvalidErr(err)
		}

		items = append(items, item)
	}

	if err := extendRecordSchema(db, items); err != nil {
		log.Debugln(err)
		response.Err = oderr.ErrDatabaseSchemaMigrationFailed
		return
	}

	results := make([]responseItem, 0, length)
	for i := range items {
		item := &items[i]
		record := &item.record

		var result responseItem
		if item.Err() {
			result = newResponseItemErr(item.record.ID.String(), item.err)
		} else if err := db.Save(record); err != nil {
			log.WithFields(log.Fields{
				"record": record,
				"err":    err,
			}).Debugln("failed to save record")

			result = newResponseItemErr(
				record.ID.String(),
				oderr.NewResourceSaveFailureErrWithStringID("record", record.ID.String()),
			)
		} else {
			result = newResponseItem((*transportRecord)(record))
		}

		results = append(results, result)
	}

	response.Result = results
}

type recordSaveItem struct {
	m      map[string]interface{}
	record oddb.Record
	err    oderr.Error
}

func (item *recordSaveItem) Err() bool {
	return item.err != nil
}

func newRecordSaveItem(mapI interface{}) recordSaveItem {
	return recordSaveItem{m: mapI.(map[string]interface{})}
}

func extendRecordSchema(db oddb.Database, items []recordSaveItem) error {
	recordSchemaMergerMap := map[string]schemaMerger{}
	for i := range items {
		recordType := items[i].record.ID.Type
		merger, ok := recordSchemaMergerMap[recordType]
		if !ok {
			merger = newSchemaMerger()
			recordSchemaMergerMap[recordType] = merger
		}

		if !items[i].Err() {
			merger.Extend(deriveRecordSchema(items[i].record.Data))
		}
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
    "database_id": "private",
    "ids": ["1004", "1005"]
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
		record := transportRecord{}
		if err := db.Get(recordID, (*oddb.Record)(&record)); err != nil {
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
			results[i] = newResponseItem(&record)
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
    "database_id": "private",
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

	results, err := db.Query(&query)
	if err != nil {
		response.Err = oderr.ErrDatabaseOpenFailed
		return
	}
	defer results.Close()

	records := []responseItem{}
	for results.Scan() {
		record := transportRecord(results.Record())
		records = append(records, newResponseItem(&record))
	}

	if err != nil {
		response.Err = oderr.ErrDatabaseQueryFailed
		return
	}

	response.Result = records
}

/*
RecordDeleteHandler is dummy implementation on delete Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:delete",
    "access_token": "validToken",
    "database_id": "_private",
    "ids": ["EA6A3E68-90F3-49B5-B470-5FFDB7A0D4E8"]
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

	results := []responseItem{}
	for i, recordID := range recordIDs {
		record := transportRecord{}
		if err := db.Get(recordID, (*oddb.Record)(&record)); err != nil {
			if err == oddb.ErrRecordNotFound {
				results[i] = newResponseItemErr(
					recordID.String(),
					oderr.ErrRecordNotFound,
				)
			} else {
				results[i] = newResponseItemErr(
					record.ID.String(),
					oderr.NewResourceDeleteFailureErrWithStringID("record", record.ID.String()),
				)
			}
		} else {
			results[i] = newResponseItem(&record)
		}
	}

	response.Result = results
	return
}

type compRecordID struct {
	kind string
	id   string
}

func (rid compRecordID) ID() string {
	return rid.kind + "/" + rid.id
}
