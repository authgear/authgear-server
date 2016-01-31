package handler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

// SchemaRenameHandler handles the action of renaming column
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
	schema, err := db.FetchSchema(recordType)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}
	log.Debugf("%+v\n", schema)

	fields := []map[string]string{}
	for key, value := range schema {
		field := map[string]string{"name": key, "type": value.Type.String()}
		fields = append(fields, field)
	}
	results := map[string]interface{}{
		"record_types": map[string]interface{}{
			recordType: map[string]interface{}{
				"fields": fields,
			},
		},
	}

	response.Result = results
}

// SchemaDeleteHandler handles the action of deleting column
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
	schema, err := db.FetchSchema(recordType)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
		return
	}

	fields := []map[string]string{}
	for key, value := range schema {
		field := map[string]string{"name": key, "type": value.Type.String()}
		fields = append(fields, field)
	}
	results := map[string]interface{}{
		"record_types": map[string]interface{}{
			recordType: map[string]interface{}{
				"fields": fields,
			},
		},
	}

	response.Result = results
}

// SchemaCreateHandler handles the action of creating new columns
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

		log.Debugf("%s\n%+v\n", recordType, recordSchema)
		err := db.Extend(recordType, recordSchema)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.IncompatibleSchema, err.Error())
			return
		}
	}

	/*
		if err := db.DeleteSchema(recordType, columnName); err != nil {
			response.Err = skyerr.NewError(skyerr.ResourceNotFound, err.Error())
			return
		}
		schema, err := db.FetchSchema(recordType)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
			return
		}

		fields := []map[string]string{}
		for key, value := range schema {
			field := map[string]string{"name": key, "type": value.Type.String()}
			fields = append(fields, field)
		}
		results := map[string]interface{}{
			"record_types": map[string]interface{}{
				recordType: map[string]interface{}{
					"fields": fields,
				},
			},
		}

		response.Result = results
	*/
}
