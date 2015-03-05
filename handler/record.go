package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"

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

func (r transportRecord) MarshalJSON() ([]byte, error) {
	// NOTE(limouren): if there is a better way to shallow copy a map,
	// do let me know
	object := map[string]interface{}{}
	for k, v := range r.Data {
		object[k] = v
	}
	object["_id"] = r.Key
	object["_type"] = r.Type

	return json.Marshal(object)
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
	id, ok := m["_id"].(string)
	if !ok {
		return errors.New(`record/json: required field "_id" not found`)
	}
	r.Key = id
	delete(m, "_id")

	t, ok := m["_type"].(string)
	if !ok {
		return errors.New(`record/json: required field "_type" not found`)
	}
	r.Type = t
	delete(m, "_type")

	r.Data = m

	return nil
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
		response.Result = oderr.New(oderr.RequestInvalidErr, "invalid request: write is not allowed")
		return
	}

	db := payload.Database
	recordMaps, ok := payload.Data["records"].([]interface{})
	if !ok {
		response.Result = oderr.New(oderr.RequestInvalidErr, "invalid request: expected list of records")
		return
	}

	length := len(recordMaps)

	records := make([]transportRecord, length, length)
	results := make([]interface{}, length, length)
	for i := range records {
		r := recordMaps[i].(map[string]interface{})
		if err := records[i].InitFromMap(r); err != nil {
			results[i] = oderr.New(oderr.RequestInvalidErr, "invalid request: "+err.Error())
		}
	}

	for i := range records {
		_, fail := results[i].(error)
		if !fail {
			if err := db.Save((*oddb.Record)(&records[i])); err != nil {
				results[i] = oderr.New(oderr.PersistentStorageErr, "persistent error: failed to save record")
			} else {
				results[i] = records[i]
			}
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
	var (
		records []oddb.Record
	)
	records = append(records, oddb.Record{
		Type: "abc",
		Key:  "abc:uuid",
	})
	log.Println("RecordFetchHandler")
	response.Result = records
	return
}

/*
RecordQueryHandler is dummy implementation on fetching Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:query",
    "access_token": "validToken",
    "database_id": "private"
}
EOF
*/
func RecordQueryHandler(payload *router.Payload, response *router.Response) {
	db := payload.Database

	recordType, _ := payload.Data["record_type"].(string)
	if recordType == "" {
		response.Result = oderr.New(oderr.RequestInvalidErr, "recordType cannot be empty")
		return
	}

	results, err := db.Query("", recordType)
	if err != nil {
		response.Result = oderr.New(oderr.UnknownErr, "failed to open database")
		return
	}
	defer results.Close()

	records := []transportRecord{}
	record := oddb.Record{}

	// needs a better abstraction here
	err = results.Next(&record)
	for err == nil {
		records = append(records, transportRecord(record))
		err = results.Next(&record)
	}

	// query failed
	if err != io.EOF {
		response.Result = oderr.New(oderr.UnknownErr, "failed to query records")
		return
	}

	response.Result = records
}

type deleteResponse struct {
	ID   string `json:"_id,omitempty"`
	Type string `json:"_type,omitempty"`
	Code string `json:"_code,omitempty"`
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
		response.Result = oderr.New(oderr.RequestInvalidErr, "invalid request: write is not allowed")
		return
	}

	db := payload.Database

	recordIDs, ok := payload.Data["ids"].([]interface{})
	if !ok {
		response.Result = oderr.New(oderr.RequestInvalidErr, "invalid request: expected list of ids")
		return
	}
	results := []deleteResponse{}
	for i := range recordIDs {
		ID, ok := recordIDs[i].(string)
		if !ok {
			results = append(results, deleteResponse{
				ID,
				"_error",
				"ID_FORMAT_ERROR", // FIXME: Dummy
			})
		} else if err := db.Delete(ID); err != nil {
			results = append(results, deleteResponse{
				ID,
				"_error",
				"NOT_FOUND", // FIXME: Dummy
			})
		}
	}
	response.Result = results
	log.Println("RecordDeleteHandler")
	return
}
