package handler

import (
	log "github.com/Sirupsen/logrus"
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

	recordTypes, err := db.FetchRecordTypes()
	if err != nil {
		return nil, err
	}
	for _, recordType := range recordTypes {
		schema, err := db.FetchSchema(recordType)
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
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SchemaRenameHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SchemaRenameHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaRenameHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("%+v", payload)

	recordType, okType := payload.Data["record_type"].(string)
	oldName, okOldName := payload.Data["item_name"].(string)
	newName, okNewName := payload.Data["new_name"].(string)
	if !okType || !okOldName || !okNewName {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "data in the specified request is invalid")
		return
	}

	db := payload.Database

	if err := db.RenameSchema(recordType, oldName, newName); err != nil {
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
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SchemaDeleteHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SchemaDeleteHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaDeleteHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("%+v", payload)

	recordType, okType := payload.Data["record_type"].(string)
	columnName, okColumnName := payload.Data["item_name"].(string)
	if !okType || !okColumnName {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "data in the specified request is invalid")
		return
	}

	db := payload.Database

	if err := db.DeleteSchema(recordType, columnName); err != nil {
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
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SchemaCreateHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SchemaCreateHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaCreateHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("%+v", payload)

	db := payload.Database

	recordTypes, ok := payload.Data["record_types"].(map[string]interface{})
	if !ok {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect map of record schema")
		return
	}

	for recordType, val := range recordTypes {
		rec, ok := val.(map[string]interface{})
		if !ok {
			response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect map of record schema")
			return
		}

		fields, ok := rec["fields"].([]interface{})
		if !ok {
			response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect list of fields")
			return
		}

		recordSchema := make(skydb.RecordSchema)
		for _, tmpField := range fields {
			field, ok := tmpField.(map[string]interface{})
			if !ok {
				response.Err = skyerr.NewError(skyerr.InvalidArgument, "unexpected field structure")
				return
			}

			fieldName, ok := field["name"].(string)
			if !ok {
				response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect string of field name")
				return
			}

			fieldTypeStr, ok := field["type"].(string)
			if !ok {
				response.Err = skyerr.NewError(skyerr.InvalidArgument, "expect string of field name")
				return
			}

			fieldType, err := skydb.SimpleNameToFieldType(fieldTypeStr)
			if err != nil {
				response.Err = skyerr.NewError(skyerr.InvalidArgument, err.Error())
				return
			}

			recordSchema[fieldName] = fieldType
		}

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
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SchemaFetchHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SchemaFetchHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SchemaFetchHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("%+v", payload)

	db := payload.Database

	results, err := prepareSchemaResponse(db)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	response.Result = results
}
