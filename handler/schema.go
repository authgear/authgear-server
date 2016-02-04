package handler

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

// prepareSchemaResponse fetchs the schema of all record types
//{
//    "record_types": {
//        "note": {
//            "fields": [
//                {"name": "content", "type": "string"},
//                {"name": "noteOrder", "type": "number"}
//            ]
//        }
//    }
//}
func prepareSchemaResponse(db skydb.Database) (map[string]map[string]interface{}, error) {
	results := map[string]map[string]interface{}{
		"record_types": map[string]interface{}{},
	}
	rt := results["record_types"]

	recordTypes, err := db.GetRecordTypes()
	if err != nil {
		return nil, err
	}
	for _, recordType := range recordTypes {
		schema, err := db.GetSchema(recordType)
		if err != nil {
			return nil, err
		}
		fields := []map[string]string{}
		for key, value := range schema {
			field := map[string]string{"name": key, "type": value.ToSimpleName()}
			fields = append(fields, field)
		}

		rt[recordType] = map[string]interface{}{
			"fields": fields,
		}
	}

	return results, nil
}

/*
SchemaRenameHandler handles the action of renaming column
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"access_token":"ee41c969-cc1f-422b-985d-ddb2217b90f8",
	"action":"schema:rename",
	"database_id":"_public",
	"record_type":"student",
	"item_type":"field",
	"item_name":"score",
	"new_name":"exam_score"
}
EOF
*/
type SchemaRenameHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *SchemaRenameHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
	}
}

func (h *SchemaRenameHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

type schemaRenamePayload struct {
	RecordType string `mapstructure:"record_type"`
	OldName    string `mapstructure:"item_name"`
	NewName    string `mapstructure:"new_name"`
}

func (payload *schemaRenamePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *schemaRenamePayload) Validate() skyerr.Error {
	if payload.RecordType == "" || payload.OldName == "" || payload.NewName == "" {
		return skyerr.NewError(skyerr.InvalidArgument, "data in the specified request is invalid")
	}
	return nil
}

func (h *SchemaRenameHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &schemaRenamePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database

	if err := db.RenameSchema(payload.RecordType, payload.OldName, payload.NewName); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
	}

	results, err := prepareSchemaResponse(db)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	response.Result = results
}

/*
SchemaDeleteHandler handles the action of deleting column
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"access_token":"ee41c969-cc1f-422b-985d-ddb2217b90f8",
	"action":"schema:delete",
	"database_id":"_public",
	"record_type":"student",
	"item_type":"field",
	"item_name":"score"
}
EOF
*/
type SchemaDeleteHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *SchemaDeleteHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
	}
}

func (h *SchemaDeleteHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

type schemaDeletePayload struct {
	RecordType string `mapstructure:"record_type"`
	ColumnName string `mapstructure:"item_name"`
}

func (payload *schemaDeletePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *schemaDeletePayload) Validate() skyerr.Error {
	if payload.RecordType == "" || payload.ColumnName == "" {
		return skyerr.NewError(skyerr.InvalidArgument, "data in the specified request is invalid")
	}
	return nil
}

func (h *SchemaDeleteHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &schemaDeletePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database

	if err := db.DeleteSchema(payload.RecordType, payload.ColumnName); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
		return
	}

	results, err := prepareSchemaResponse(db)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	response.Result = results
}

/*
SchemaCreateHandler handles the action of creating new columns
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"access_token":"ee41c969-cc1f-422b-985d-ddb2217b90f8",
	"action":"schema:create",
	"database_id":"_public",
	"record_types":{
		"student": {
			"fields":[
				{"name": "age", 	"type": "number"},
				{"name": "nickname" "type": "string"}
			]
		}
	}
}
EOF
*/
type SchemaCreateHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *SchemaCreateHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
	}
}

func (h *SchemaCreateHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

type schemaCreatePayload struct {
	RawSchemas map[string]struct {
		Fields []struct {
			Name     string `mapstructure:"name"`
			TypeName string `mapstructure:"type"`
		} `mapstructure:"fields"`
	} `mapstructure:"record_types"`

	Schemas map[string]skydb.RecordSchema
}

func (payload *schemaCreatePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	payload.Schemas = make(map[string]skydb.RecordSchema)
	for recordType, schema := range payload.RawSchemas {
		payload.Schemas[recordType] = make(skydb.RecordSchema)
		for _, field := range schema.Fields {
			var err error
			payload.Schemas[recordType][field.Name], err = skydb.SimpleNameToFieldType(field.TypeName)
			if err != nil {
				return skyerr.NewError(skyerr.InvalidArgument, "unexpected field type")
			}
		}
	}

	return payload.Validate()
}

func (payload *schemaCreatePayload) Validate() skyerr.Error {
	for _, schema := range payload.Schemas {
		for fieldName := range schema {
			if strings.HasPrefix(fieldName, "_") {
				return skyerr.NewError(skyerr.InvalidArgument, "attempts to create reserved field")
			}
		}
	}
	return nil
}

func (h *SchemaCreateHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("%+v\n", rpayload)

	payload := &schemaCreatePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database

	for recordType, recordSchema := range payload.Schemas {
		err := db.Extend(recordType, recordSchema)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.IncompatibleSchema, err.Error())
			return
		}
	}

	results, err := prepareSchemaResponse(db)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}
	response.Result = results
}

/*
SchemaFetchHandler handles the action of returing information of record schema
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"access_token":"ee41c969-cc1f-422b-985d-ddb2217b90f8",
	"action":"schema:fetch",
	"database_id":"_public"
}
EOF
*/
type SchemaFetchHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *SchemaFetchHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
	}
}

func (h *SchemaFetchHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaFetchHandler) Handle(rpayload *router.Payload, response *router.Response) {
	db := rpayload.Database

	results, err := prepareSchemaResponse(db)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	response.Result = results
}
