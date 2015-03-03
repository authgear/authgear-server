package handler

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/oursky/ourd/auth"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/router"
)

// RecordHandler declares the interface of a handler that works with records
type RecordHandler func(*recordPayload, *router.Response, oddb.Database)

// RecordService provides a collection of handlers to
// handle oddb.Record related operations on an oddb.Database.
type RecordService struct {
	auth.TokenStore
}

// injectRecordHandler returns a router.Handler that has a proper
// public / private database injected into RecordHandler according to
// the payload
func (s RecordService) injectRecordHandler(recordHandler RecordHandler) router.Handler {
	return func(rpayload *router.Payload, response *router.Response) {
		payload := newRecordPayload(rpayload)

		if !payload.IsValidDB() {
			response.Result = NewError(MissingDatabaseIDErr, "Invalid Database ID")
			return
		}

		var db oddb.Database
		token := auth.Token{}
		if payload.IsPublicDB() {
			if !payload.IsReadOnly() {
				if err := s.TokenStore.Get(payload.AccessToken(), &token); err != nil {
					response.Result = NewError(InvalidAccessTokenErr, "Invalid access token")
					return
				}
			}
			db = payload.DBConn.PublicDB()
		} else { // if a request doesn't ask for public DB, then it is private DB
			if err := s.TokenStore.Get(payload.AccessToken(), &token); err != nil {
				response.Result = NewError(InvalidAccessTokenErr, "Invalid access token")
				return
			}

			db = payload.DBConn.PrivateDB(token.UserInfoID)
		}

		recordHandler(&payload, response, db)
	}
}

// RecordFetchHandler returns a router.Handler that fetches a record.
func (s RecordService) RecordFetchHandler() router.Handler {
	return s.injectRecordHandler(RecordSaveHandler)
}

// RecordSaveHandler returns a router.Handler that saves a record.
func (s RecordService) RecordSaveHandler() router.Handler {
	return s.injectRecordHandler(RecordSaveHandler)
}

// RecordDeleteHandler returns a router.Handler that deletes a record.
func (s RecordService) RecordDeleteHandler() router.Handler {
	return s.injectRecordHandler(RecordSaveHandler)
}

// RecordQueryHandler returns a router.Handler that queries records.
func (s RecordService) RecordQueryHandler() router.Handler {
	return s.injectRecordHandler(RecordSaveHandler)
}

// recordPayload is the input parameter in RecordHandler
type recordPayload struct {
	*router.Payload
	DatabaseID string
}

func newRecordPayload(payload *router.Payload) recordPayload {
	databaseID, _ := payload.Data["database_id"].(string)
	return recordPayload{
		Payload:    payload,
		DatabaseID: databaseID,
	}
}

func (p recordPayload) IsValidDB() bool {
	return p.DatabaseID == "_public" || p.DatabaseID == "_private"
}

func (p recordPayload) IsPublicDB() bool {
	return p.DatabaseID == "_public"
}

func (p recordPayload) IsReadOnly() bool {
	action := p.RouteAction()
	return action == "record:fetch" || action == "record:query"
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
  -d @- http://localhost:3000/ <<EOF
{
    "action": "record:save",
    "access_token": "validToken",
    "database_id": "private"
}
EOF
*/
func RecordSaveHandler(payload *recordPayload, response *router.Response, db oddb.Database) {
	log.Println("RecordSaveHandler")
	return
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
func RecordFetchHandler(payload *recordPayload, response *router.Response, db oddb.Database) {
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
func RecordQueryHandler(payload *recordPayload, response *router.Response, db oddb.Database) {
	log.Println("RecordQueryHandler")
	return
}

/*
RecordDeleteHandler is dummy implementation on delete Records
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "redord:delete",
    "access_token": "validToken",
    "database_id": "private"
}
EOF
*/
func RecordDeleteHandler(payload *recordPayload, response *router.Response, db oddb.Database) {
	log.Println("RecordDeleteHandler")
	return
}
