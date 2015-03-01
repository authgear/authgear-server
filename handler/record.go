package handler

import (
	"log"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/router"
)

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
func RecordSaveHandler(payload *router.Payload, response *router.Response) {
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
func RecordDeleteHandler(payload *router.Payload, response *router.Response) {
	log.Println("RecordDeleteHandler")
	return
}
