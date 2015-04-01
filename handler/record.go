package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

// recordPayload is the input parameter in RecordHandler
type recordPayload router.Payload

func (p *recordPayload) IsWriteAllowed() bool {
	return p.UserInfo == nil
}

// transportRecord override JSON serialization and deserialization of
// oddb.Record
type transportRecord oddb.Record

func (r transportRecord) ID() string {
	return r.Type + "/" + r.Key
}

func (r transportRecord) MarshalJSON() ([]byte, error) {
	// NOTE(limouren): marshalling of type/key is delegated to responseItem
	return json.Marshal(r.Data)
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
	delete(m, "_id")

	ss := strings.SplitN(rawID, "/", 2)
	if len(ss) == 1 {
		return fmt.Errorf(`record/json: "_id" should be of format '{type}/{id}', got %#v`, rawID)
	}

	recordType, id := ss[0], ss[1]

	r.Key = id
	r.Type = recordType
	r.Data = m

	return nil
}

type responseItem struct {
	id     string
	record *transportRecord
	err    oderr.Error
}

func newResponseItem(record *transportRecord) responseItem {
	return responseItem{
		id:     record.ID(),
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
		buf.Write([]byte(`error",`))
		i = item.err
	} else if item.record != nil {
		buf.Write([]byte(`record",`))
		i = item.record
	} else {
		panic("inconsistent state: both err and record is nil")
	}

	bodyBytes, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	if bodyBytes[0] != '{' {
		return nil, fmt.Errorf("first char of embedded json != {: %v", string(bodyBytes))
	} else if bodyBytes[len(bodyBytes)-1] != '}' {
		return nil, fmt.Errorf("last char of embedded json != }: %v", string(bodyBytes))
	}
	buf.Write(bodyBytes[1:])
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
*/
func RecordSaveHandler(payload *router.Payload, response *router.Response) {
	if (*recordPayload)(payload).IsWriteAllowed() {
		response.Err = oderr.ErrWriteDenied
		return
	}

	db := payload.Database
	recordMaps, ok := payload.Data["records"].([]interface{})
	if !ok {
		response.Err = oderr.NewRequestInvalidErr(errors.New("expected list of record"))
		return
	}

	length := len(recordMaps)

	records := make([]transportRecord, length, length)
	results := make([]responseItem, length, length)
	for i := range records {
		r := recordMaps[i].(map[string]interface{})
		if err := records[i].InitFromMap(r); err != nil {
			response.Err = oderr.NewRequestInvalidErr(err)
			return
		} else if err := db.Save((*oddb.Record)(&records[i])); err != nil {
			results[i] = newResponseItemErr(
				records[i].ID(),
				oderr.NewResourceSaveFailureErrWithStringID("record", records[i].ID()),
			)
		} else {
			results[i] = newResponseItem(&records[i])
		}
	}

	response.Result = results
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
	recordIDs := make([]compRecordID, length, length)
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

		recordIDs[i] = compRecordID{ss[0], ss[1]}
	}

	db := payload.Database

	results := make([]responseItem, length, length)
	for i, recordID := range recordIDs {
		record := transportRecord{}
		if err := db.Get(recordID.id, (*oddb.Record)(&record)); err != nil {
			if err == oddb.ErrRecordNotFound {
				results[i] = newResponseItemErr(
					recordID.ID(),
					oderr.ErrRecordNotFound,
				)
			} else {
				results[i] = newResponseItemErr(
					recordID.ID(),
					oderr.NewResourceFetchFailureErr("record", recordID.ID()),
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

	records := []transportRecord{}
	for results.Scan() {
		records = append(records, transportRecord(results.Record()))
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
	if (*recordPayload)(payload).IsWriteAllowed() {
		response.Err = oderr.ErrWriteDenied
		return
	}

	db := payload.Database

	interfaces, ok := payload.Data["ids"].([]interface{})
	if !ok {
		response.Err = oderr.NewRequestInvalidErr(errors.New("expected list of id"))
		return
	}

	length := len(interfaces)
	recordIDs := make([]compRecordID, length, length)
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

		recordIDs[i] = compRecordID{ss[0], ss[1]}
	}

	results := []responseItem{}
	for i, recordID := range recordIDs {
		record := transportRecord{}
		if err := db.Get(recordID.id, (*oddb.Record)(&record)); err != nil {
			if err == oddb.ErrRecordNotFound {
				results[i] = newResponseItemErr(
					recordID.ID(),
					oderr.ErrRecordNotFound,
				)
			} else {
				results[i] = newResponseItemErr(
					record.ID(),
					oderr.NewResourceDeleteFailureErrWithStringID("record", record.ID()),
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
